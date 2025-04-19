# 🐘 Mammoth Server

**Mammoth Server** é um serviço leve escrito em Go para **comparação de esquemas de banco de dados** e **geração de scripts de migração**. Ideal para times que desejam automatizar o versionamento e a evolução de schemas com controle, segurança e facilidade.

## 🚀 Funcionalidades

- 🔍 Comparação entre dois esquemas de banco de dados (fonte e destino)
- 🧠 Detecção de diferenças em:
  - Tabelas (adicionadas/removidas)
  - Colunas (nome, tipo, nulabilidade, chave primária)
  - Índices
  - Chaves estrangeiras
- 🛠 Geração de scripts de migração (DDL) automaticamente
- 🌐 API RESTful com endpoints para integração
- 📦 Suporte atual: PostgreSQL (MySQL em breve)

## 📦 Instalação

```bash
git clone https://github.com/seu-usuario/mammoth-server.git
cd mammoth.server
go build -o mammoth
```

## 🧪 Uso rápido

```bash
./mammoth
```
