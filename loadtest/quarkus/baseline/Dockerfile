# Use the official k6 image as a base image
FROM grafana/k6:latest

# Set the working directory
WORKDIR /app

# Copy the k6 script into the container
#COPY quarkus-baseline-breakpoint.js .
COPY quarkus-baseline-v2.js .

# Command to run the k6 script
#CMD ["run", "quarkus-baseline-breakpoint.js"]
CMD ["run", "quarkus-baseline-v2.js"]
