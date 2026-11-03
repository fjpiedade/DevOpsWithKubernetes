# Log output

Esta aplicação é composta por dois containers no mesmo Pod:

- `log-writer`: mantém `log.txt` com timestamp UTC e UUID.
- `log-reader`: expõe `GET /` e combina o conteúdo de `log.txt` com o contador
  persistente `pingpong.txt`.

O writer recupera o UUID de `log.txt` quando reinicia, em vez de gerar outro.
Ambos os ficheiros ficam no PVC partilhado com o serviço Ping-pong.
