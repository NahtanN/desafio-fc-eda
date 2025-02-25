## Dependências

- Docker
- Docker Compose

## Para rodar o projeto

Execute o comando abaixo para rodar o projeto

```bash
docker compose up -d
```

As migrations, seeds e tópicos do kafka serão executadas automaticamente.

Os usuários abaixo serão criados, você pode usar qualquer um deles para testar as transações.

Usuário 1:

```json
{
  "id": "d77d5d9b-6959-4637-8aaf-4c677e2fa83e",
  "name": "foo",
  "email": "foo@bar",
  "account": {
    "id": "1fe35ef4-bbc7-4a23-80c4-48c966dbbc5f",
    "balance": 1000
  }
}
```

Usuário 2:

```json
{
  "id": "611d3cef-f1c6-4fa7-a821-0b5edec151d6",
  "name": "bar",
  "email": "bar@foo",
  "account": {
    "id": "4be4234a-156a-4679-8a8a-35d8f1293502",
    "balance": 1000
  }
}
```

As chamadas http estão no diretório `api`.
