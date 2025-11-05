# Git Hooks

В проекте используются **pre-commit хуки** для автоматической проверки и форматирования кода перед каждым коммитом.

## Установка

Настрой путь к хукам один раз:

```bash
git config core.hooksPath deployments/git-hooks
