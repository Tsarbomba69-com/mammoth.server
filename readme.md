# 🐘 Mammoth Server

**Mammoth Server** is a lightweight service written in Go for **database schema comparison** and **migration script generation**. Ideal for teams looking to automate schema versioning and evolution with control, security, and ease.

## 🚀 Features

- 🔍 Comparison between two database schemas (source and target)
- 🧠 Detection of differences in:
  - Tables (added/removed)
  - Columns (name, type, nullability, primary key)
  - Indexes
  - Foreign keys
- 🛠 Automatic generation of migration scripts (DDL)
- 🌐 RESTful API with endpoints for integration
- 📦 Current support: PostgreSQL (MySQL coming soon)

## 📦 Installation

```bash
git clone https://github.com/Tsarbomba69-com/mammoth.server.git
cd mammoth.server
go build -o mammoth
```

## 🧪 Quick Start

```bash
./mammoth
```
