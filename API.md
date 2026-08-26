Autenticação

A autenticação utiliza JWT.

Rotas protegidas devem enviar:

Authorization: Bearer {token}

O JWT contém:

{
  "user_id": 1,
  "email": "professor@escola.com",
  "role": "teacher"
}

Roles

Existem atualmente duas roles:

    teacher
    student

Auth
Cadastrar professor

Permite o cadastro manual de professores.

POST /auth/register-teacher

Body

{
  "email": "professor@escola.com",
  "password": "senha123",
  "name": "Carlos Professor"
}

O usuário criado recebe:

role = teacher

Cadastrar aluno

Somente professores podem cadastrar alunos.

POST /auth/register-student
Authorization: Bearer {teacher_token}

Body

{
  "email": "aluno@escola.com",
  "password": "senha123",
  "name": "João Silva"
}

O usuário criado recebe:

role = student

Login

POST /auth/login

Body

{
  "email": "professor@escola.com",
  "password": "senha123"
}

Resposta

{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}

O token deve ser enviado nas próximas requisições protegidas.
Eventos
Criar evento

Somente professores.

POST /events
Authorization: Bearer {teacher_token}

Body

{
  "title": "Feira de Ciências",
  "description": "Feira anual de projetos científicos.",
  "place": "Auditório da Escola",
  "category": "Ciência",
  "date": "2027-01-20T14:00:00Z",
  "end_date": "2027-01-20T16:00:00Z",
  "max_participants": 50
}

Regras

    date deve ser anterior a end_date.
    max_participants deve ser maior que zero.
    O sistema gera automaticamente um token de check-in para o evento.
    O evento pode ser gerenciado por professores.

Listar eventos

Professores e alunos autenticados.

GET /events
Authorization: Bearer {token}

Retorna somente eventos futuros.

Os eventos são ordenados pela data de início, do mais próximo para o mais distante.
Consultar evento

Professores e alunos autenticados.

GET /events/:id
Authorization: Bearer {token}

Resposta

{
  "data": {
    "id": 1,
    "title": "Feira de Ciências",
    "description": "Feira anual de projetos científicos.",
    "place": "Auditório da Escola",
    "category": "Ciência",
    "date": "2027-01-20T14:00:00Z",
    "end_date": "2027-01-20T16:00:00Z",
    "max_participants": 50,
    "registered_count": 35,
    "available_slots": 15
  }
}

available_slots representa:

max_participants - registered_count

Nunca será retornado um valor negativo.
Atualizar evento

Somente professores.

PUT /events/:id
Authorization: Bearer {teacher_token}

Body

{
  "title": "Feira de Ciências 2027",
  "description": "Feira anual de projetos científicos.",
  "place": "Ginásio da Escola",
  "category": "Ciência",
  "date": "2027-01-20T14:00:00Z",
  "end_date": "2027-01-20T17:00:00Z",
  "max_participants": 100
}

O número máximo de participantes não pode ser menor que o número atual de inscritos.
Excluir evento

Somente professores.

DELETE /events/:id
Authorization: Bearer {teacher_token}

Ao excluir um evento, também são excluídos:

    inscrições;
    registros de presença;
    certificados.

A exclusão é realizada dentro de uma transaction.
Inscrições

Somente alunos podem se inscrever em eventos.
Inscrever-se

POST /events/:id/register
Authorization: Bearer {student_token}

Não é necessário enviar body.
Exemplo

curl -X POST http://localhost:8080/events/1/register \
  -H "Authorization: Bearer $STUDENT_TOKEN"

Regras

    O aluno pode se inscrever apenas uma vez.
    O evento precisa possuir vagas.
    O evento precisa estar disponível para inscrição.
    O aluno precisa estar autenticado como student.

Se o aluno já estiver inscrito:

409 Conflict

Se o evento estiver lotado:

409 Conflict

Cancelar inscrição

DELETE /events/:id/register
Authorization: Bearer {student_token}

Não é necessário enviar body.

O aluno pode cancelar sua inscrição.
Check-in / Presença

O sistema possui dois mecanismos para confirmação de presença:

    QR Code exibido pelo professor.
    Botão de confirmação dentro do aplicativo do aluno.

Ambos utilizam a mesma regra de negócio no backend.

A presença só pode ser registrada uma vez.
Janela de check-in

O check-in fica disponível:

30 minutos antes do início do evento

até:

o horário de término do evento

Exemplo:

Evento:
14:00 → 16:00

Check-in:
13:30 → 16:00

Fora dessa janela o check-in é rejeitado.
QR Code do evento
Consultar QR Code

Somente professores.

GET /events/:id/check-in
Authorization: Bearer {teacher_token}

Essa rota é utilizada pela tela de check-in do professor.
Resposta

{
  "data": {
    "event_id": 1,
    "title": "Feira de Ciências",
    "date": "2027-01-20T14:00:00Z",
    "end_date": "2027-01-20T16:00:00Z",
    "check_in_starts_at": "2027-01-20T13:30:00Z",
    "check_in_ends_at": "2027-01-20T16:00:00Z",
    "qr_code": "data:image/png;base64,iVBORw0KGgo...",
    "total_registered": 35,
    "total_present": 12
  }
}

qr_code é uma imagem PNG em formato Data URL.

O frontend pode exibir diretamente:

<img src="data:image/png;base64,..." />

Tela de check-in do professor

O professor pode manter a tela aberta durante o evento.

A tela deve exibir:

Feira de Ciências

QR CODE

Inscritos: 35
Presentes: 12

O frontend pode consultar periodicamente:

GET /events/:id/check-in

para atualizar a quantidade de presentes.

Não é necessário WebSocket para o MVP.
Check-in do aluno
Confirmar presença

POST /events/:id/check-in
Authorization: Bearer {student_token}

A rota pode ser utilizada de duas maneiras.
Opção 1 — Pelo botão do aplicativo

O aluno abre o evento no aplicativo e seleciona:

Confirmar presença

Nesse caso, não é necessário enviar token.
Body

{}

O backend verifica:

    usuário autenticado;
    usuário é aluno;
    aluno está inscrito;
    evento está dentro da janela de check-in;
    aluno ainda não fez check-in.

Opção 2 — Pelo QR Code

O aluno escaneia o QR Code exibido pelo professor.

O QR Code contém uma URL semelhante a:

http://localhost:3000/check-in/1/{check_in_token}

O frontend extrai o token e envia:

POST /events/1/check-in
Authorization: Bearer {student_token}
Content-Type: application/json

{
  "token": "{check_in_token}"
}

Nesse caso, além das validações normais, o backend valida se o token pertence ao evento.
Resultado do check-in

Quando a presença é confirmada, o sistema cria:

    EventAttendance
    Certificate

Essas duas operações são realizadas dentro da mesma transaction.
Resposta

{
  "message": "Attendance confirmed successfully",
  "data": {
    "attendance": {
      "id": 1,
      "event_id": 1,
      "student_id": 2,
      "checked_in_at": "2027-01-20T13:45:00Z"
    },
    "certificate": {
      "id": 1,
      "code": "CERT-550e8400-e29b-41d4-a716-446655440000"
    }
  }
}

Erros de check-in
Aluno não está inscrito

403 Forbidden

{
  "error": "Forbidden",
  "message": "Student is not registered for this event"
}

Check-in fora do horário

409 Conflict

{
  "error": "Conflict",
  "message": "Check-in is not available at this time"
}

Check-in duplicado

409 Conflict

{
  "error": "Conflict",
  "message": "Student already checked in"
}

QR Code inválido

403 Forbidden

{
  "error": "Forbidden",
  "message": "Invalid check-in token"
}

Certificados

Um certificado é criado automaticamente quando o aluno confirma sua presença.

O certificado funciona como comprovante de participação presencial no evento.
Baixar certificado

Somente o aluno proprietário do certificado.

GET /certificates/:id/pdf
Authorization: Bearer {student_token}

A resposta é um arquivo:

Content-Type: application/pdf

Exemplo:

curl http://localhost:8080/certificates/1/pdf \
  -H "Authorization: Bearer $STUDENT_TOKEN" \
  -o certificado.pdf

O aluno não pode acessar o certificado pertencente a outro aluno.
Conteúdo do certificado

O PDF contém informações como:

    nome do aluno;
    título do evento;
    descrição;
    categoria;
    local;
    data do evento;
    data/hora da presença;
    código de autenticidade.

Exemplo:

CERTIFICADO

Certificamos que

JOÃO SILVA

participou presencialmente do evento
"Feira de Ciências",

realizado em 20/01/2027,
no Auditório da Escola.

Categoria: Ciência
Data da presença: 20/01/2027 13:45

Código de autenticidade:
CERT-550e8400-...

Validação pública do certificado

A validação não exige autenticação.

GET /certificates/verify/:code

Exemplo

curl http://localhost:8080/certificates/verify/CERT-550e8400-...

Certificado válido

{
  "valid": true,
  "data": {
    "certificate_code": "CERT-550e8400-...",
    "student": {
      "name": "João Silva"
    },
    "event": {
      "title": "Feira de Ciências",
      "description": "Feira anual de projetos científicos.",
      "category": "Ciência",
      "place": "Auditório da Escola",
      "date": "2027-01-20T14:00:00Z",
      "end_date": "2027-01-20T16:00:00Z"
    },
    "issued_at": "2027-01-20T13:45:00Z"
  }
}

Certificado inválido

404 Not Found

{
  "valid": false,
  "error": "Not Found"
}

Resumo das rotas
Método	Endpoint	Permissão	Descrição
POST	/auth/login	Público	Login
POST	/auth/register-teacher	Público	Cadastro de professor
POST	/auth/register-student	Professor	Cadastro de aluno
GET	/events	Autenticado	Lista eventos futuros
POST	/events	Professor	Criar evento
GET	/events/:id	Autenticado	Detalhes do evento
PUT	/events/:id	Professor	Atualizar evento
DELETE	/events/:id	Professor	Excluir evento
POST	/events/:id/register	Aluno	Inscrever-se
DELETE	/events/:id/register	Aluno	Cancelar inscrição
GET	/events/:id/check-in	Professor	Consultar QR Code
POST	/events/:id/check-in	Aluno	Confirmar presença
GET	/certificates/:id/pdf	Aluno	Baixar certificado
GET	/certificates/verify/:code	Público	Validar certificado
Fluxo completo
Professor

Login
  ↓
Criar evento
  ↓
Aguardar evento
  ↓
Abrir tela de check-in
  ↓
QR Code é exibido
  ↓
Acompanhar quantidade de presentes

Aluno — inscrição

Login
  ↓
Consultar eventos
  ↓
Abrir evento
  ↓
Inscrever-se
  ↓
Aguardar evento

Aluno — presença via QR Code

Evento começa
  ↓
Professor exibe QR Code
  ↓
Aluno escaneia
  ↓
Frontend abre check-in
  ↓
Aluno confirma
  ↓
Backend valida
  ↓
Presença registrada
  ↓
Certificado gerado

Aluno — presença pelo aplicativo

Evento começa
  ↓
Aluno abre o evento
  ↓
Botão "Confirmar presença"
  ↓
Backend valida
  ↓
Presença registrada
  ↓
Certificado gerado

Certificado

Check-in confirmado
  ↓
Certificado criado
  ↓
Aluno baixa PDF
  ↓
Código de autenticidade
  ↓
Qualquer pessoa pode validar

Regras de negócio atuais

    Professores podem criar, editar e excluir eventos.
    Alunos podem visualizar eventos futuros.
    Somente alunos podem se inscrever.
    Um aluno pode se inscrever apenas uma vez em cada evento.
    Alunos podem cancelar suas inscrições.
    Eventos possuem limite máximo de participantes.
    Eventos não podem ser atualizados para um limite inferior ao número atual de inscritos.
    O check-in começa 30 minutos antes do evento.
    O check-in termina quando o evento termina.
    Somente alunos inscritos podem confirmar presença.
    Um aluno pode confirmar presença apenas uma vez.
    A presença pode ser confirmada pelo QR Code ou pelo aplicativo.
    O QR Code do evento é disponibilizado somente para professores autenticados.
    A confirmação de presença gera automaticamente um certificado.
    Presença e certificado são criados na mesma transaction.
    Um aluno só pode baixar seus próprios certificados.
    Certificados podem ser validados publicamente através do código de autenticidade.
    Ao excluir um evento, suas inscrições, presenças e certificados também são excluídos.
