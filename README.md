# Limpeza Segura GUI 2.0

Utilitário gráfico para Windows 11 escrito em Go. Limpa somente arquivos temporários e caches selecionados.

## Dados preservados

- Senhas salvas
- Cookies e sessões
- Histórico
- Favoritos
- Extensões
- Perfis dos navegadores

## Base Baboo v3.2 e proteções adicionais

Esta versão incorpora as categorias úteis do script Baboo v3.2, mas protege por
padrão os dados de sites, IndexedDB, Service Worker, armazenamento web, cache de
certificados, bancos do Explorer e dados persistentes do TeamViewer. Esses itens
não são apagados porque podem causar logout, falhas de páginas ou reconstruções
lentas.

Lixeira, WebCache, logs, dumps e caches pesados de aplicativos ficam desmarcados
por padrão. Somente temporários do usuário atual são processados; o programa não
tenta alterar permissões nem limpar perfis de outros usuários.

## Compilar

1. Extraia a pasta para `H:\projects\Go\LimpezaSeguraGUI`.
2. Execute `build.bat`.
3. Aguarde o download das dependências.
4. O executável será criado em `dist\LimpezaSeguraGUI.exe`.

O script está configurado para usar o Go instalado em `C:\Program Files\Go`.

> Importante: o projeto usa `github.com/tc-hib/go-winres`, compatível com Go
> Modules. O manifesto ativa os controles comuns v6 exigidos pela Walk. Isso evita
> o erro `TTM_ADDTOOL failed`. Execute novamente o `build.bat`; ele remove
> automaticamente recursos antigos antes de gerar o novo EXE.

## Desenvolvimento manual

```bat
cd /d H:\projects\Go\LimpezaSeguraGUI
"C:\Program Files\Go\bin\go.exe" mod tidy
"C:\Program Files\Go\bin\go.exe" generate
"C:\Program Files\Go\bin\go.exe" build -ldflags="-H windowsgui" -o dist\LimpezaSeguraGUI.exe .
```
