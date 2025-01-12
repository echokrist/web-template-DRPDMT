# Base Stage: Common setup for all environments
FROM denoland/deno:latest AS base

# Set the working directory inside the container
WORKDIR /frontend

# Set the DENO_DIR for persistent caching
ENV DENO_DIR=/deno-dir

# Ensure the directory exists and Deno has permission to write to it
RUN mkdir -p $DENO_DIR && chmod 777 $DENO_DIR

# Copy project files into the container
COPY . .

# Expose the port your app runs on
EXPOSE 8000

########################################################

# DEVELOPMENT
FROM base AS development

# Install dependencies and prepare the application for development
RUN deno cache main.ts

# Command to run the development server
CMD ["deno", "task", "dev"]

########################################################

# BUILD Stage: Used for staging and production
FROM base AS build

# Run the production task to bundle the application
# Assumes a "production" task is defined in your `deno.json`
RUN deno task production

########################################################

# STAGING
FROM denoland/deno:2.1.5 AS staging

# Set the working directory
WORKDIR /frontend

# Copy the bundled application from the build stage
COPY --from=build /frontend/app.bundle.js .

# Command to run the bundled application for staging
CMD ["deno", "run", "--allow-net", "app.bundle.js"]

########################################################

# PRODUCTION
FROM denoland/deno:2.1.5 AS production

# Set the working directory
WORKDIR /frontend

# Copy the bundled application from the build stage
COPY --from=build /frontend/app.bundle.js .

# Command to run the bundled application for production
CMD ["deno", "run", "--allow-net", "app.bundle.js"]
