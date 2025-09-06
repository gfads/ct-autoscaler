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
const loadDuration = 10
const baseVUs = 100
const baseRandom = 9

export const options = {
    noVUConnectionReuse: true,
    noConnectionReuse: true,
    scenarios: {
        /*load: {
            // load (rate * duration) records into DB
            executor: 'constant-arrival-rate',
            rate: loadRate,
            timeUnit: '1s',
            duration: `${loadDuration}s`,
            preAllocatedVUs: 100,
            maxVUs: 500,
            exec: 'load',
        },*/
        experiment: {
            executor: 'ramping-vus',
            //startTime: `${loadDuration+30}s`, // wait for load db be executed
            startVUs: baseVUs,
            //timeUnit: '1s',
            //preAllocatedVUs: 0, // Initial number of VUs
            //maxVUs: 240, // Maximum number of VUs
            gracefulRampDown: '0s',
            stages: [
                { target: 2*baseVUs+randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 2*baseVUs+randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 2*baseVUs+randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 2*baseVUs +randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 3*baseVUs +randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 3*baseVUs +randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 3*baseVUs+randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 3*baseVUs+randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 1*baseVUs+randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 1*baseVUs+randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 1*baseVUs+randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 1*baseVUs+randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 2*baseVUs+randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 2*baseVUs+randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 2*baseVUs+randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 2*baseVUs+randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 3*baseVUs+randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 3*baseVUs+randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 3*baseVUs+randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 3*baseVUs+randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 2*baseVUs+randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 2*baseVUs+randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 2*baseVUs+randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
                { target: 2*baseVUs+randomIntBetween(baseRandom*-0.1, baseRandom*0.1), duration: '1m' },
            ]
        },
    },
};

export function load() {
    const listResponse = http.get(url);
    check(listResponse, {
        'load LIST status is 200': (r) => r.status === 200,
    });

    const listResponseBody = listResponse.json();
    if (listResponseBody.length < loadRate * loadDuration) {
        const postName = `fruit_${__ITER}_${__VU}`;
        const postResponse = http.post(url, JSON.stringify({ name: postName }), { headers: postHeaders, tags: { name: 'post' } });
        check(postResponse, {
            'Load POST status is 201': (r) => r.status === 201,
        });
    }

}

export default function workload() {
    const postName = `fruit_${__VU * 2}`;

    const postResponse = http.post(url, JSON.stringify({ name: postName }), { headers: postHeaders, tags: { name: 'post' } });

    check(postResponse, {
        'POST status is 201': (r) => r.status === 201,
    });

    const postResponseBody = postResponse.json();
    const id = postResponseBody.id;

    const getResponse = http.get(`${url}/${id}`, { tags: { name: 'get' } });
    check(getResponse, {
        'GET status is 200': (r) => r.status === 200,
    });

    const putName = `fruit_${__VU * 2 - 1}`;

    const putResponse = http.put(`${url}/${id}`, JSON.stringify({ name: putName }), { headers: postHeaders, tags: { name: 'put' } });
    check(putResponse, {
        'PUT status is 200': (r) => r.status === 200,
    });

    for (let i = 0; i < 5; i++) {
        const listResponse = http.get(url);
        check(listResponse, {
            'LIST status is 200': (r) => r.status === 200,
        });

        const listResponseBody = listResponse.json();
        if (listResponseBody.length > 0) {
            const randomRecordId = listResponseBody[Math.floor(Math.random() * listResponseBody.length)].id;
            const randomRecordGetResponse = http.get(`${url}/${randomRecordId}`, { tags: { name: 'random' } });
            check(randomRecordGetResponse, {
                'GET random record status is 200': (r) => r.status === 200,
            });
        }
    }

    //const deleteResponse = http.del(`${url}/${id}`, null, { tags: { name: 'delete' } });
    //check(deleteResponse, {
    //    'DELETE status is 204': (r) => r.status === 204,
   // });
}