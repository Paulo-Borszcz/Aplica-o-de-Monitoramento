# Aplicação de Monitoramento

E aí, galera! Bem-vindos ao meu projeto de monitoramento de sistema. Passei um bom tempo desenvolvendo isso, então vou compartilhar o que aprendi no processo. Vamos nessa!

## O que esse negócio faz?

Basicamente, essa aplicação coleta uma porrada de informações do sistema, incluindo hardware, software, rede e performance. Depois, ela criptografa tudo isso e manda pra um servidor. Massa, né?

## Estrutura do Projeto

O projeto tá organizado assim:

- `main.go`: O coração da aplicação. Coordena tudo.
- `hardware/`: Coleta info de CPU, memória, disco, GPU, placa-mãe, BIOS e dispositivos USB.
- `software/`: Pega dados do sistema operacional, kernel, apps instalados, processos rodando e serviços.
- `network/`: Lida com interfaces de rede, conexões, DNS, IP público e info avançada de rede.
- `performance/`: Monitora uso de CPU, memória, I/O de disco e rede, carga do sistema e temperaturas.
- `utils/`: Funções utilitárias, tipo criptografia e leitura de arquivos INI.

## Como funciona?

1. A aplicação lê o arquivo `config.ini` pra pegar o endereço do servidor e a chave de criptografia.
2. Coleta todas as informações do sistema.
3. Transforma tudo em um objeto JSON maneiro.
4. Criptografa esse JSON usando AES-256-CFB.
5. Codifica o resultado em Base64 URL.
6. Manda tudo pro servidor via POST.

## Detalhes da Criptografia

Usei AES-256-CFB pra criptografar os dados. A chave é lida do arquivo `config.ini`. É importante manter esse arquivo seguro e não compartilhar a chave!

## Coleta de Dados

### Hardware

- CPU: modelo, cores, threads, frequência, temperatura, uso
- Memória: total, usada, livre, porcentagem de uso
- Disco: dispositivo, tipo, total, usado, livre, porcentagem de uso
- GPU: modelo, memória, temperatura, uso
- Placa-mãe: fabricante, modelo, número de série
- BIOS: fornecedor, versão, data de lançamento
- Dispositivos USB: nome, ID do fornecedor, ID do produto, número de série

### Software

- Sistema Operacional: nome, versão, arquitetura, hostname
- Kernel: versão
- Aplicativos instalados: nome, versão, data de instalação
- Processos em execução: nome, PID, uso de CPU, uso de memória
- Serviços do sistema: nome, status

### Rede

- Interfaces: nome, endereço MAC, endereços IP, status, velocidade, bytes enviados/recebidos
- Conexões: endereço local/remoto, porta, estado, processo
- Servidores DNS
- IP público
- Info avançada: latência, perda de pacotes, velocidade de download/upload, tabela de roteamento, configuração DNS, status VPN

### Performance

- Uso de CPU
- Uso de memória
- I/O de disco: bytes lidos/escritos, IOPS
- I/O de rede: bytes enviados/recebidos, pacotes enviados/recebidos
- Carga do sistema
- Temperaturas: CPU, GPU, disco

## Como usar

1. Certifique-se de ter Go instalado (usei a versão 1.20).
2. Clone o repositório.
3. Rode `go mod tidy` pra pegar as dependências.
4. Crie um arquivo `config.ini` com o seguinte conteúdo:

server_address=[http://seu-servidor.com/endpoint]
(http://seu-servidor.com/endpoint)

5. Execute `go run main.go`.

## Configuração

O arquivo `config.ini` deve conter:

- `server_address`: O endereço do servidor para onde os dados serão enviados.
- `encryption_key`: Uma chave hexadecimal de 64 caracteres (32 bytes) para criptografia AES-256.

## Observações Importantes

1. Mantenha o `config.ini` seguro! Ele contém informações sensíveis.
2. A aplicação precisa de permissões elevadas para coletar alguns dados do sistema.
3. Em sistemas Windows, alguns dados podem requerer privilégios de administrador.
4. Em sistemas Linux, você pode precisar instalar pacotes adicionais para coletar certas informações.

## Observações Finais

Esse projeto foi um desafio e tanto! Aprendi muito sobre coleta de dados do sistema, criptografia e Go. Se tiver dúvidas ou sugestões, me chama que a gente troca uma ideia. Ah, e se encontrar algum bug, me avisa, tá?

Boa sorte aí e manda ver!
