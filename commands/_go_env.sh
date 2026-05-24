# Usa Go 1.25 del toolchain si la máquina tiene 1.24 (go.mod y deps lo requieren).
# Si tienes GOTOOLCHAIN=local en tu shell, este script lo anula solo para los comandos del repo.
export GOTOOLCHAIN=auto
