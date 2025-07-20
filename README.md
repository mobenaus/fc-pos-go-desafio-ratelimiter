# fc-pos-go-desafio-ratelimiter

## Desafio
```
Objetivo: Desenvolver um rate limiter em Go que possa ser configurado para limitar o número máximo de requisições por segundo com base em um endereço IP específico ou em um token de acesso.

Descrição: O objetivo deste desafio é criar um rate limiter em Go que possa ser utilizado para controlar o tráfego de requisições para um serviço web. O rate limiter deve ser capaz de limitar o número de requisições com base em dois critérios:

Endereço IP: O rate limiter deve restringir o número de requisições recebidas de um único endereço IP dentro de um intervalo de tempo definido.
Token de Acesso: O rate limiter deve também poderá limitar as requisições baseadas em um token de acesso único, permitindo diferentes limites de tempo de expiração para diferentes tokens. O Token deve ser informado no header no seguinte formato:
API_KEY: <TOKEN>
As configurações de limite do token de acesso devem se sobrepor as do IP. Ex: Se o limite por IP é de 10 req/s e a de um determinado token é de 100 req/s, o rate limiter deve utilizar as informações do token.
Requisitos:

O rate limiter deve poder trabalhar como um middleware que é injetado ao servidor web
O rate limiter deve permitir a configuração do número máximo de requisições permitidas por segundo.
O rate limiter deve ter ter a opção de escolher o tempo de bloqueio do IP ou do Token caso a quantidade de requisições tenha sido excedida.
As configurações de limite devem ser realizadas via variáveis de ambiente ou em um arquivo “.env” na pasta raiz.
Deve ser possível configurar o rate limiter tanto para limitação por IP quanto por token de acesso.
O sistema deve responder adequadamente quando o limite é excedido:
Código HTTP: 429
Mensagem: you have reached the maximum number of requests or actions allowed within a certain time frame
Todas as informações de "limiter” devem ser armazenadas e consultadas de um banco de dados Redis. Você pode utilizar docker-compose para subir o Redis.
Crie uma “strategy” que permita trocar facilmente o Redis por outro mecanismo de persistência.
A lógica do limiter deve estar separada do middleware.
Exemplos:

Limitação por IP: Suponha que o rate limiter esteja configurado para permitir no máximo 5 requisições por segundo por IP. Se o IP 192.168.1.1 enviar 6 requisições em um segundo, a sexta requisição deve ser bloqueada.
Limitação por Token: Se um token abc123 tiver um limite configurado de 10 requisições por segundo e enviar 11 requisições nesse intervalo, a décima primeira deve ser bloqueada.
Nos dois casos acima, as próximas requisições poderão ser realizadas somente quando o tempo total de expiração ocorrer. Ex: Se o tempo de expiração é de 5 minutos, determinado IP poderá realizar novas requisições somente após os 5 minutos.
Dicas:

Teste seu rate limiter sob diferentes condições de carga para garantir que ele funcione conforme esperado em situações de alto tráfego.
Entrega:

O código-fonte completo da implementação.
Documentação explicando como o rate limiter funciona e como ele pode ser configurado.
Testes automatizados demonstrando a eficácia e a robustez do rate limiter.
Utilize docker/docker-compose para que possamos realizar os testes de sua aplicação.
O servidor web deve responder na porta 8080.
```

## Funcionamento

O funcionamento da implementação é baseada no RateLimit com o algoritmo Token Bucket, onde a cada unidade de tempo determinada é disponibilizada uma quantidade determinada de tokens em um bucket.

Sempre que for necessário acesso a algum recurso que deva ser protegido pelo rate limiter, deve ser removido um token do bucket, quando os tokens forem consumidos, o rate limiter deve recusar o acesso.

Nesse momento na verificação dos tokens restantes, também deve ser verificado a última vez que o tokens foram disponibilizados, e se o tempo for maior que o determinado, novos tokens devem ser gerados e o acesso ao recurso pode ser liberado.

A implementação do ratelimiter é implementada dentro de um MiddleWare que é anexado ao servidor Web da implementação

Esse middleware é configurado com 2 tipos de ratelimiter, um para IP e outro para TOKEN que podem ter configurações de capacidade e período distintas.

O ratelimiter de TOKEN só será utilizado em requisições que possuam o header API_KEY

O ratelimiter de IP será utilizado para as requisições que não possuem o header API_KEY


## Implementação

Nesse desafio, foi implementado um servidor web muito simples em main.go.

### Configuração
No main as configurações são carregadas via arquivo .env e variaveis de ambiente usando o framework viper

As configurações que são carregadas são as seguintes:
- *RATE_LIMIT_STRATEGY*: Tipo de persistencia que será utilizado no controle dos buckets do rate limiter, *REDIS* ou *MEMORY*.
- *REDIS_ADDR*: Caso a persistencia dos buckets seja REDIS, este parâmetro indica o endereço do servidor Redis, no formato host:porta.
- *REDIS_PASSWORD*: Caso o Redis precise de senha, ela deve ser colocada neste parâmetro.
- *REDIS_DEFAULT_DB*: Indica o banco de dados no servidor Redis, inteiro, normalmente é 0.
- *IP_RATE_LIMIT*: Capacidade do bucket do ratelimit para endereços IP.
- *IP_RATE_PERIOD*: Periodo de preenchimento do bucket (por exemplo 1s).
- *TOKEN_RATE_LIMIT*: Capacidade do bucket do ratelimit para tokens (header API_KEY).
- *TOKEN_RATE_PERIOD*: Periodo de preenchimento do bucket (por exemplo 1s).
- *WEB_SERVER_PORT*: Porta do servidor web do desafio.

### Inicialização
Após a leitura da configuração, os 2 RateLimiters (IP e TOKEN) são instanciados usando a persistencia configurada.

O MiddleWare é configurado usando os 2 RateLimiters.

O servidor Web é iniciado usando o MiddleWare de ratelimiter e um middleware de Logging.

Quando a estratégia de ratelimiter é via REDIS, a implementação executa a conexão com o servidor e testa se ele está ativo.

### MiddleWare

O middleware, incia verificando se a requisição possui o header API_KEY extraindo o seu valor.

Caso encontre um valor, o middleware utiliza o ratelimiter correspondente, usando esse valor como chave para o ratelimiter.

Caso não encontre um valor no header API_KEY, o ratelimiter de IP é utilizado, considerando a parte do IP do RemoteAddr da requisição.

O middleware vai chamar o metodo UseToken do ratelimiter adequado, que poderá retornar erro se a requisição for recursada, nesta situação, a requisição é retornada como conteudo do erro e o status 429 - Too Many Requests

Caso não retorne erro, significa que um token foi consumido e o acesso da requisição foi autorizado e a requisição pode ser encaminhada para o proximo Middleware.

### RateLimiter

A implementação do ratelimiter está em ratelimit.RateLimit (internal/ratelimit/ratelimit.go).

É definida uma Struct com uma interface de Persistencia do ratelimiter e um mutex para evitar acesso concorrente ao ratelimiter dentro da mesma instancia.

A função **NewRateLimit** cria uma nova instancia do RateLimit, com a implementação de persistencia passada por parâmetro.

Aqui tambem é implementado para a struct RateLimit a função UseToken que recebe uma key por parametro.

A logica do ratelimit que executa é a seguinte:
- Usa o mutex para isolar os acessos
- Busca o bucket da persistencia
- Verifica se o bucket deve ser preenchido
- Em caso positivo, preenche o bucket com a capacidade total
- Verifica se os tokens do bucket foram todos consumidos.
- Decrementa a quantidade de tokens do bucket
- Salva o bucket na persistencia.
Em caso de erros em alguma operação ou se os tokens forem todos consumidos, o erro retornado é com a mensagem: "you have reached the maximum number of requests or actions allowed within a certain time frame"

### RateLimiterPersistence

Essa interface determina que as implementações de persisntencia para o rate limiter devem ter os seguintes metodos:
- GetBucket(key string) (*Bucket, error)
- CheckRefill(bucket *Bucket) bool
- Refill(bucket *Bucket)
- SaveBucket(key string, bucket *Bucket) error

Estes metodos são utilizados no RateLimiter.UseToken

A persistencia deve utilizar a key como forma de segregar os ratelimits.

por exemplo, se multiplos IPs ou multiplas API_KEYs são utilizadas, o rate limit é cara cada key.

As operações de CheckRefill e Refill são delegadas para a persistencia, porque as configurações de capacidade e periodo de refill é configurado na implementação de persistencia.

### Bucket
A struct Bucket mantem o número de tokens e o momento do último enchimento do bucket.

### MemoryRateLimitPersistence

A implementação em memoria do ratelimit é muito simples, baseando a persistencia dos buckets em um Map onde a chave é a key (IP da requisição ou valor do API_KEY)

A função **NewMemoryRateLimitPersistence** inicializa uma nova persistencia configurando a capacidade e periodo para enchimento dos buckets.

**Refil(bucket)**, apenas atualiza o momento do refill e reseta os tokens do bucket para a capacidade configurada.

**GetBucket(key)** busca um bucket no mapa, ou cria um novo bucket com a capacidade configurada.

**CheckReffil(bucket)** verifica se o tempo da ultima vez que o bucket foi preenchido já passou

**SaveBucket(key, bucket)** salva o bucket no mapa em memoria na chave key

### RedisRateLimitPersistence

A implementação com persistencia do ratelimit em Redis, fazendo a persistencia dos buckets em Sets, onde para cada bucket temos 2 keys:
- "\<prefix>:refill:\<key>"
- "\<prefix>:tokens:\<key>"

Nessas keys, o prefix é determinado na criação e identifica se o bucket é para IP ou TOKEN de API, e key identifica o IP da requisição ou o conteúdo do header API_KEY.

A função **NewRedisRateLimitPersistence** inicializa uma nova persistencia configurando o context, o client Redis, o prefix, a capacidade e periodo para enchimento dos buckets.

**Refil(bucket)**, apenas atualiza o momento do refill e reseta os tokens do bucket para a capacidade configurada.

**GetBucket(key)** busca um bucket no Redis, ou cria um novo bucket com a capacidade configurada se não encontrar

**CheckReffil(bucket)** verifica se o tempo da ultima vez que o bucket foi preenchido já passou

**SaveBucket(key, bucket)** salva o bucket no Redis, os dados são salvos usando o periodo do RateLimit como tempo de expiração dos registros, não há necessidade de manter esses registros por mais tempo.

A funcão interna **fill()** é utilizada pela função GetBucket para executar a busca dos dados no Redis.

## Testes automatizados

Foram implemetados testes em Go para testar a logica do RateLimit em *internal/ratelimit/ratelimit_test.go*

```
go test -v ./internal/ratelimit
```
O resultado esperado é:
```
$ go test -v ./internal/ratelimit
=== RUN   TestUseToken_Success
--- PASS: TestUseToken_Success (0.00s)
=== RUN   TestUseToken_Refill
--- PASS: TestUseToken_Refill (0.00s)
=== RUN   TestUseToken_GetBucketError
--- PASS: TestUseToken_GetBucketError (0.00s)
=== RUN   TestUseToken_SaveBucketError
--- PASS: TestUseToken_SaveBucketError (0.00s)
=== RUN   TestUseToken_NoTokens
--- PASS: TestUseToken_NoTokens (0.00s)
=== RUN   TestRateLimitFiveRequestsMemoryPersistence
--- PASS: TestRateLimitFiveRequestsMemoryPersistence (0.00s)
PASS
ok      github.com/mobenaus/fc-pos-go-desafio-ratelimiter/internal/ratelimit    0.005s
```
## Execução via docker compose

Para executar o servidor web, junto com Redis e 2 containers para teste do ratelimit, pode ser executado o comando
```
docker compose up -d
```
Ele irá instanciar um Redis, o server (depois do redis) e mais 2 containers para testes do ratelimiter.

O ratelimit esta configurado para usar o Redis.
O ratelimit de IP é de 50 requisições a cada 1 segundo
O ratelimit de API é de 100 requisições a cada 1 segundo

O resultado dos testes para o container test-api-limit podem ser verificados visualizando o log desse container com:
```
docker compose logs test-api-limit
```
O teste utiliza a imagem rcmorano/docker-hey configurado para fazer 1000 requisições durante 2 segundos, em 2 processos concorrentes com 50 requisições por segundo em cada processo e passando um header API_KEY nas requisições.
O resultado deve ser que aproximadamente todas as requisições serão aceitas com status 200
```
test-api-limit  | Status code distribution:
test-api-limit  |   [200]       989 responses
test-api-limit  |   [429]       11 responses
```
Outro teste com o container test-ip-limit pode ser verificado visualizando os logs deste container com:
```
docker compose logs test-ip-limit
```
O teste teve as mesmas configurações, porem sem passar o header API_KEY, o resultado deve ser aproximadamente metade das requisições são aceitas com o status 200 enquando o resto é retornado com o status 429.
```
test-ip-limit  | Status code distribution:
test-ip-limit  |   [200]        500 responses
test-ip-limit  |   [429]        500 responses
```

Para parar os serviço deve ser utilizado:
```
docker compose down
```