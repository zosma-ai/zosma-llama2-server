# zosma-llama2-server
Task Queue Manager for distributed LLaMA-2 inference worker network based on [asynq](https://github.com/hibiken/asynq) and [redis](https://github.com/redis/redis).

Worker is created by integrating workers/worker with [LLaMA 2 refrence inference](https://github.com/rampenke/zosma-llama2-worker)

Client applications, for example [web server](./clients/webserver/server.go) uses clinet interface to uses the distributed inference network.

