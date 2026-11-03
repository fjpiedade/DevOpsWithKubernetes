# Ping-pong

Serviço HTTP que incrementa e guarda o número de pedidos em `pingpong.txt`.
O ficheiro é montado a partir do PVC partilhado, pelo que o contador sobrevive
à recriação do Pod.

Endpoint: `GET /pingpong` (porta interna 9901).
