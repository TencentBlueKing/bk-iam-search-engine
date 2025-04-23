debug: true

server:
  host: 127.0.0.1
  port: 9001

  readTimeout: 60
  writeTimeout: 60
  idleTimeout: 180

sentry:
  enable: false
  dsn: ""

superAppCode: "bk_iam,iam"

storage:
  path: "./"

index:
  elasticsearch:
    indexName: iam_policy
    addresses:
      - https://local.ipv6-dual.bktencent.com:9200
    username: "elastic"
    password: "123456"
    maxRetries: 3
    tls:                            # tls配置
      enabled: true                 # 是否开启tls
      certCaFile: "your CA file"    # ca证书路径

backend:
    addr: "http://127.0.0.1:9000"
    authorization:
      appCode: "bk_iam"
      appSecret: "appSecret"

redis:
  id: "mq"
  type: "standalone"
  addr: "dev.ipv6-dual.bktencent.com:6389"
  password: ""
  db: 0
  poolSize: 160
  dialTimeout: 3
  readTimeout: 1
  writeTimeout: 1
  sentinelAddr: ""
  masterName: "mymaster"
  sentinelPassword: "" # TODO: waiting for the replace var
  tls:
    enabled: true
    certCaFile: "your CA file"

redisKeys:
  - id: "delete_queue_key"
    key: "bk_iam:deleted_policy"


logger:
  system:
    level: debug
    writer: os
    settings: {name: stdout}
  sync:
    level: debug
    writer: os
    settings: {name: stdout}
    # level: info
    # writer: file
    # settings: {name: iam_search_engine_sync.log, size: 100, backups: 10, age: 7, path: ./}
  es:
    level: debug
    writer: os
    settings: {name: stdout}
    # level: info
    # writer: file
    # settings: {name: iam_search_engine_es.log, size: 100, backups: 10, age: 7, path: ./}
  api:
    level: info
    writer: file
    settings: {name: iam_search_engine_api.log, size: 100, backups: 10, age: 7, path: ./}
  component:
    level: info
    writer: file
    settings: {name: iam_search_engine_component.log, size: 100, backups: 10, age: 7, path: ./}
