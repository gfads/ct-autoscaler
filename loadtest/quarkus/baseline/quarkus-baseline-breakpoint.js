import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';
import { Counter } from 'k6/metrics';

const requestRate = __ENV.REQUEST_RATE ? parseInt(__ENV.REQUEST_RATE) : 20; // Number of requests per second
const host = __ENV.HOST || 'localhost';
const port = __ENV.PORT || 8080;
const url = `http://${host}:${port}/fruits`;

const postHeaders = { 'Content-Type': 'application/json' };

// number of records loaded into DB is loadRate X loadDuration
// play around these numbers to adjust the initial size of db
const loadRate = 10;
const loadDuration = 30

export const options = {
    thresholds: {
        //http_req_failed: ['rate<0.1'], // http errors should be less than 1%
        //http_req_duration: ['p(95)<200'], // 95% of requests should be below 200ms
        http_req_failed: [
            {
                threshold: 'rate<0.03',
                abortOnFail: true,
                delayAbortInterval: '5s'
            }
            
        ]
      },
      stages: [
        { duration: '10m', target: 2000 }],
    noVUConnectionReuse: true,
    noConnectionReuse: true,
    executor: 'ramping-arrival-rate',
   /* scenarios: {
        load: {
            // load (rate * duration) records into DB
            executor: 'constant-arrival-rate',
            rate: loadRate,
            timeUnit: '1s',
            duration: `${loadDuration}s`,
            preAllocatedVUs: 100,
            maxVUs: 500,
            exec: 'load',
        },*/
        /*experiment: {
            executor: 'constant-arrival-rate',
            //startTime: `${loadDuration+30}s`, // wait for load db be executed
            //startRate: requestRate,
            //startRate: 200,
            timeUnit: '1s',
            preAllocatedVUs: 10, // Initial number of VUs
            maxVUs: 550, // Maximum number of VUs
            rate: 300,
            duration: '10m'
        },
    },*/
};

export function load() {
    const listResponse = http.get(url);
    check(listResponse, {
        'load LIST status is 200': (r) => r.status !== 500,
    });

    const listResponseBody = listResponse.json();
    if (listResponseBody.length < loadRate * loadDuration) {
        const postName = `fruit_${__ITER}_${__VU}`;
        const postResponse = http.post(url, JSON.stringify({ name: postName }), { headers: postHeaders, tags: { name: 'post' } });
        check(postResponse, {
            'Load POST status is 201': (r) => r.status !== 500,
        });
    }

}

export default function workload() {
   
    const postName = `fruit_${__VU * 2}`;

    const postResponse = http.post(url, JSON.stringify({ name: postName }), { headers: postHeaders, tags: { name: 'post' } });

    check(postResponse, {
        'POST status is 201': (r) => r.status !== 500,
    });

    const postResponseBody = postResponse.json();
    const id = postResponseBody.id;

    const getResponse = http.get(`${url}/${id}`, { tags: { name: 'get' } });
    check(getResponse, {
        'GET status is 200': (r) => r.status !== 500,
    });

    const putName = `fruit_${__VU * 2 - 1}`;

    const putResponse = http.put(`${url}/${id}`, JSON.stringify({ name: putName }), { headers: postHeaders, tags: { name: 'put' } });
    check(putResponse, {
        'PUT status is 200': (r) => r.status !== 500,
    });

    for (let i = 0; i < 5; i++) {
        const listResponse = http.get(url);
        check(listResponse, {
            'LIST status is 200': (r) => r.status !== 500,
        });

        const listResponseBody = listResponse.json();
        if (listResponseBody.length > 0) {
            const randomRecordId = listResponseBody[Math.floor(Math.random() * listResponseBody.length)].id;
            const randomRecordGetResponse = http.get(`${url}/${randomRecordId}`, { tags: { name: 'random' } });
            check(randomRecordGetResponse, {
                'GET random record status is 200': (r) => r.status !== 500,
            });
        }
    }

    const deleteResponse = http.del(`${url}/${id}`, null, { tags: { name: 'delete' } });
    check(deleteResponse, {
        'DELETE status is 204': (r) => r.status !== 500,
    });
}
