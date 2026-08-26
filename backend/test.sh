#!/bin/bash

set -e

API="http://localhost:8080"

echo "========================================"
echo "   TESTE DA API - EVENTOS ESCOLARES"
echo "========================================"

echo ""
echo "1. Cadastrando professor..."

curl -s -X POST "$API/auth/register-teacher" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "professor@escola.com",
    "password": "senha123",
    "name": "Carlos Professor"
  }'

echo ""
echo "OK"

echo ""
echo "2. Login do professor..."

TEACHER_RESPONSE=$(curl -s -X POST "$API/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "professor@escola.com",
    "password": "senha123"
  }')

echo "$TEACHER_RESPONSE"

TEACHER_TOKEN=$(echo "$TEACHER_RESPONSE" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')

if [ -z "$TEACHER_TOKEN" ]; then
    echo "ERRO: não foi possível obter o token do professor"
    exit 1
fi

echo "Token do professor obtido."

echo ""
echo "3. Professor criando evento..."

EVENT_RESPONSE=$(curl -s -X POST "$API/events" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TEACHER_TOKEN" \
  -d '{
    "title": "Feira de Ciências",
    "description": "Feira anual de ciências da escola",
    "place": "Auditório Principal",
    "category": "Ciência",
    "date": "2026-12-20T14:00:00Z",
    "max_participants": 2
  }')

echo "$EVENT_RESPONSE"

EVENT_ID=$(echo "$EVENT_RESPONSE" | sed -n 's/.*"id":\([0-9]*\).*/\1/p')

if [ -z "$EVENT_ID" ]; then
    echo "ERRO: não foi possível obter o ID do evento"
    exit 1
fi

echo "Evento criado: ID=$EVENT_ID"

echo ""
echo "4. Listando eventos futuros..."

curl -s -X GET "$API/events" \
  -H "Authorization: Bearer $TEACHER_TOKEN"

echo ""

echo ""
echo "5. Buscando evento..."

curl -s -X GET "$API/events/$EVENT_ID" \
  -H "Authorization: Bearer $TEACHER_TOKEN"

echo ""

echo ""
echo "6. Cadastrando aluno..."

curl -s -X POST "$API/auth/register-student" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TEACHER_TOKEN" \
  -d '{
    "email": "aluno1@escola.com",
    "password": "senha123",
    "name": "Maria Santos"
  }'

echo ""

echo ""
echo "7. Login do aluno..."

STUDENT_RESPONSE=$(curl -s -X POST "$API/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "aluno1@escola.com",
    "password": "senha123"
  }')

echo "$STUDENT_RESPONSE"

STUDENT_TOKEN=$(echo "$STUDENT_RESPONSE" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')

if [ -z "$STUDENT_TOKEN" ]; then
    echo "ERRO: não foi possível obter o token do aluno"
    exit 1
fi

echo "Token do aluno obtido."

echo ""
echo "8. Aluno consultando eventos..."

curl -s -X GET "$API/events" \
  -H "Authorization: Bearer $STUDENT_TOKEN"

echo ""

echo ""
echo "9. Aluno se inscrevendo no evento..."

curl -s -X POST "$API/events/$EVENT_ID/register" \
  -H "Authorization: Bearer $STUDENT_TOKEN"

echo ""

echo ""
echo "10. Verificando vagas..."

curl -s -X GET "$API/events/$EVENT_ID" \
  -H "Authorization: Bearer $STUDENT_TOKEN"

echo ""

echo ""
echo "11. Tentando inscrição duplicada..."

curl -s -X POST "$API/events/$EVENT_ID/register" \
  -H "Authorization: Bearer $STUDENT_TOKEN"

echo ""

echo ""
echo "12. Cadastrando segundo aluno..."

curl -s -X POST "$API/auth/register-student" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TEACHER_TOKEN" \
  -d '{
    "email": "aluno2@escola.com",
    "password": "senha123",
    "name": "João Silva"
  }'

echo ""

echo ""
echo "13. Login do segundo aluno..."

STUDENT2_RESPONSE=$(curl -s -X POST "$API/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "aluno2@escola.com",
    "password": "senha123"
  }')

echo "$STUDENT2_RESPONSE"

STUDENT2_TOKEN=$(echo "$STUDENT2_RESPONSE" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')

echo ""
echo "14. Segundo aluno se inscrevendo..."

curl -s -X POST "$API/events/$EVENT_ID/register" \
  -H "Authorization: Bearer $STUDENT2_TOKEN"

echo ""

echo ""
echo "15. Verificando evento lotado..."

curl -s -X GET "$API/events/$EVENT_ID" \
  -H "Authorization: Bearer $STUDENT2_TOKEN"

echo ""

echo ""
echo "16. Segundo aluno tenta criar evento..."

curl -s -X POST "$API/events" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $STUDENT2_TOKEN" \
  -d '{
    "title": "Evento indevido",
    "description": "Teste",
    "place": "Sala 1",
    "category": "Teste",
    "date": "2026-12-25T14:00:00Z",
    "max_participants": 10
  }'

echo ""

echo ""
echo "17. Primeiro aluno cancelando inscrição..."

curl -s -X DELETE "$API/events/$EVENT_ID/register" \
  -H "Authorization: Bearer $STUDENT_TOKEN"

echo ""

echo ""
echo "18. Verificando vaga liberada..."

curl -s -X GET "$API/events/$EVENT_ID" \
  -H "Authorization: Bearer $STUDENT_TOKEN"

echo ""

echo ""
echo "19. Professor atualizando evento..."

curl -s -X PUT "$API/events/$EVENT_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TEACHER_TOKEN" \
  -d '{
    "title": "Feira de Ciências 2026",
    "description": "Feira anual de ciências - edição 2026",
    "place": "Ginásio",
    "category": "Ciência",
    "date": "2026-12-21T15:00:00Z",
    "max_participants": 5
  }'

echo ""

echo ""
echo "20. Professor excluindo evento..."

curl -s -X DELETE "$API/events/$EVENT_ID" \
  -H "Authorization: Bearer $TEACHER_TOKEN"

echo ""

echo ""
echo "21. Tentando buscar evento excluído..."

curl -s -X GET "$API/events/$EVENT_ID" \
  -H "Authorization: Bearer $TEACHER_TOKEN"

echo ""

echo ""
echo "========================================"
echo "        TESTES FINALIZADOS"
echo "========================================"
