# Replay: medir o autodirector em demos

Roda a lógica de câmera do app (pacote `gui/director`) em cima de demos do CS2,
simulando o GSI (atualização a cada 0,1 s, loop do director a cada 0,5 s e
0,25 s de atraso por troca), e mede quantas kills a câmera pegou.

## Uso

```powershell
cd tools\replay
go mod tidy          # só na primeira vez
go run . "C:\...\csgo\rushfy-1.dem" "C:\...\csgo\rushfy-2.dem"
```

Saída por demo e total:

- **camera on killer at the kill**: câmera no matador no momento da kill
- **1s / 2s / 3s before**: câmera já no matador antes (mede previsão)
- **kill on screen**: matador ou vítima na câmera
- **switches per live minute**: trocas por minuto de round ao vivo

`missed_kills.csv` lista cada kill perdida com o `tick` (use `demo_gototick <tick>`
no CS2 para ver o lance).

Rounds com mais de 3 minutos "ao vivo" são tratados como bug da demo e ignorados.

## Opções

| Opção | O que faz |
|---|---|
| `-legacy` | usa a lógica antiga (sem previsão), para comparar |
| `-pred "margin=0.05,dwell=3"` | muda as regras de troca da previsão |
| `-modeldir pasta` | usa `pasta\<demo>.json` como modelo (teste com modelo que não viu a demo) |
| `-dataset dados.csv` | gera os dados de treino do modelo |
| `-aimcheck` | confere se a direção da mira está sendo lida certo |
| `-v` | mostra as decisões do director |

## Retreinar o modelo com novas demos

```powershell
go run . -dataset dados.csv <demos...>
python train.py dados.csv modelos      # precisa de pandas e scikit-learn
go run . -modeldir modelos <demos...>  # mede com modelos que não viram cada demo
copy modelos\model.go ..\..\gui\director\model.go   # atualiza o modelo do app
```
