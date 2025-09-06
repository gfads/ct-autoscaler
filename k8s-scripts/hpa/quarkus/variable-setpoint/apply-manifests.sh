#!/bin/bash

# Verifica se há argumentos suficientes
if [ "$#" -lt 3 ]; then
  echo "Uso: $0 intervalo_segundos tempo_total_segundos manifesto1.yaml [manifesto2.yaml ...]"
  exit 1
fi

# Lê os argumentos
INTERVALO="$1"
TEMPO_TOTAL="$2"
shift 2
MANIFESTOS=("$@")

# Nome do arquivo de log (com timestamp de início)
LOGFILE="aplicacao_manifestos_$(date '+%Y%m%d_%H%M%S').log"

# Verifica se os manifestos existem
for MANIFESTO in "${MANIFESTOS[@]}"; do
  if [ ! -f "$MANIFESTO" ]; then
    echo "Erro: Arquivo '$MANIFESTO' não encontrado!"
    exit 1
  fi
done

echo "Iniciando aplicação sequencial dos manifestos a cada $INTERVALO segundos, por $TEMPO_TOTAL segundos..."
echo "Logs serão salvos em: $LOGFILE"
echo "Pressione Ctrl+C para interromper antecipadamente."

INICIO=$(date +%s)
INDEX=0
TOTAL_MANIFESTOS=${#MANIFESTOS[@]}

while true; do
  AGORA=$(date +%s)
  ELAPSED=$((AGORA - INICIO))

  if [ "$ELAPSED" -ge "$TEMPO_TOTAL" ]; then
    echo "Tempo total de execução atingido. Encerrando."
    break
  fi

  # Seleciona o manifesto da vez
  MANIFESTO="${MANIFESTOS[$INDEX]}"
  TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
  echo "[$TIMESTAMP] Aplicando $MANIFESTO..." | tee -a "$LOGFILE"

  if kubectl apply -f "$MANIFESTO" >> "$LOGFILE" 2>&1; then
    echo "[$TIMESTAMP] SUCESSO - $MANIFESTO" >> "$LOGFILE"
  else
    echo "[$TIMESTAMP] ERRO    - $MANIFESTO" >> "$LOGFILE"
  fi

  # Avança para o próximo manifesto, com loop circular
  INDEX=$(( (INDEX + 1) % TOTAL_MANIFESTOS ))

  echo "Aguardando $INTERVALO segundos..." | tee -a "$LOGFILE"
  sleep "$INTERVALO"
done

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Execução finalizada." | tee -a "$LOGFILE"
