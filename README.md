# Signal-Art-Creator

This is a service for running [create.signal.art](https://create.signal.art).

## Building

This project is consists of a backend service written in Go, and front-end in
TypeScript. In order to run both on a server, the front-end code has to be built
first:
```sh
cd web
npm install
npm run build
cd -
```

...and then embedded into the go executable:
```
go build -o ./artd ./cmd/artd
```

## Running

The resulting executable file can be started to run full service locally:
```sh
./artd config.yaml
```

## Running against staging/production server

It is also possible to run the front-end code against an already deployed
backend service:
```sh
cd web
npm install
npm run dev
```

This will start a local web server on https://localhost:5173/, and a proxy
server that would map `http://localhost:5173/api/*` to a deployed backend.

The URL of the backend can be configured by editing the "server" section of
`./web/vite.config.ts`. As an example, here is a configuration for running the
front-end against the production server:
```js
server: {
  proxy: {
    '/api/socket': {
      secure: true,
      target: 'wss://create.signal.art',
      changeOrigin: true,
      headers: {
        origin: 'https://create.signal.art',
      },
    },
    '/api': {
      secure: true,
      target: 'https://create.signal.art',
      changeOrigin: true,
    },
  },
},
```

## Configuration
A sample configuration:
```yaml
bucket: "<s3BucketName>"
region: "<region>"
authSecret: "<sharedAuthSecret>"
uploadURL: "<s3UploadUrl>"
redisClusterURI: "<redisClusterUri>"
rateLimiter:
  bucketName: "<ratelimiterBucketName>"
  bucketSize: 10
  leakRatePerMinute: 1
provisioning:
  pubSubPrefix: "<prefix>"
  origin: "https://create.signal.art"
```

