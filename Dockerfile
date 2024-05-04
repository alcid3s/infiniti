# syntax=docker/dockerfile:1

#Constant arguments for the build.
ARG GO_VERSION=1.22.2
ARG PORT=9000

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS final

# Constant paths for the custom packages made (if more packages needed, add here).
ENV MAINPATH=/src/main
ENV AUDIOPIPELINEPATH=/src/audiopipeline

# Set current directory.
WORKDIR /src

# Copy project into image.
COPY * ./

# Download dependencies for Audiopipeline Package.
WORKDIR ${AUDIOPIPELINEPATH}
RUN go mod download -x

# Download dependencies for Main Package.
WORKDIR ${MAINPATH}
RUN go mod download -x

# Create a build for the application.
RUN CGO_ENABLED=0 GOOS=linux go build -o /src/bin/infiniti

# Expose port 8080 to the outside world.
EXPOSE ${PORT}

# Command ran when the container is started.
CMD [ "/src/bin/infiniti" ]