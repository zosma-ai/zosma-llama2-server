# zosma-llama2-server

**Repository Description: zosma-llama2-server**

This repository is designed to manage a distributed network of LLaMA-2 inference workers through a Task Queue Manager utilizing `asynq` and Redis. It sets up a scalable framework where each worker node is integrated with the LLaMA 2 reference inference model to process tasks efficiently. The system is engineered to support client applications, such as web servers, enabling them to utilize the distributed inference network seamlessly. This infrastructure allows for high-throughput and low-latency language model inference, catering to applications requiring advanced natural language processing capabilities. The setup is optimized for performance and scalability, providing a robust backend for handling extensive LLaMA-2 inference operations.

## Refere the following repos for additional information
- [asynq](https://github.com/hibiken/asynq) 
- [redis](https://github.com/redis/redis).
- [LLaMA 2 refrence inference](https://github.com/rampenke/zosma-llama2-worker)
- [web server](./clients/webserver/server.go) uses clinet interface to uses the distributed inference network.

## Build worker
```
docker build -t <image-tag>  -f Dockerfile.worker .
```
Example:
```
docker build -t mcntech/zosma-llama2-worker  -f Dockerfile.worker .
```

## Build and run webserver
```
docker build -t <image-tag>  -f Dockerfile.webserver .
```
Example:
```
docker build -t mcntech/zosma-llama2-webserver  -f Dockerfile.webserver .
docker run -it --env-file .env mcntech/zosma-llama2-webserver
```