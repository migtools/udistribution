package def

const Config = `version: 0.1
log:
  level: debug
  fields:
    service: registry
    environment: development
  hooks:
    - type: mail
      disabled: true
      levels:
        - panic
      options:
        smtp:
          addr: mail.example.com:25
          username: mailuser
          password: password
          insecure: true
        from: sender@example.com
        to:
          - errors@example.com
storage:
    delete:
      enabled: true
    cache:
        blobdescriptor: redis
    filesystem:
        rootdirectory: /var/lib/registry
    maintenance:
        uploadpurging:
            enabled: false
http:
    addr: :5000
    debug:
        addr: :5001
        prometheus:
            disabled: true
            path: /metrics
    headers:
        X-Content-Type-Options: [nosniff]
redis:
  addr: localhost:6379
  pool:
    maxidle: 16
    maxactive: 64
    idletimeout: 300s
  dialtimeout: 10ms
  readtimeout: 10ms
  writetimeout: 10ms
health:
  storagedriver:
    enabled: true
    interval: 10s
    threshold: 3
`
