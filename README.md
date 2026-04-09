<h1 align="center"> 
 A Rota Das Coisas
</h1>

---

<details>
  <summary><h2> Descrição do Projeto</h2></summary>

O projeto **A Rota Das Coisas** é um sistema distribuído desenvolvido em **Go (Golang)** que atua como um **Middleware de Integração IoT**. Ele foi projetado para resolver o problema de alto acoplamento e gargalos de rede em ecossistemas de Internet das Coisas (IoT). 

A solução atua como um intermediário inteligente, desacoplando quem produz a informação (dispositivos físicos como sensores) de quem a consome (aplicações cliente. Ele recebe dados simultâneos, gerencia o estado global, aplica regras automáticas de negócio e repassa comandos críticos de forma otimizada.

</details>

---

<details>
  <summary><h2> Contexto e Problema</h2></summary>

Em arquiteturas IoT tradicionais (ponto-a-ponto), dispositivos físicos com recursos limitados de processamento e memória precisam gerenciar conexões diretas com múltiplas aplicações simultaneamente. Se um cliente deseja apresentar dados de um sensor em um painel, gravar em um banco e acionar um alarme, o sensor precisa cuidar de todas essas conexões. Isso gera sobrecarga, lentidão e travamentos.

Além disso, aplicações diferentes apresentam necessidades de tráfego diferentes que sistemas convencionais não distinguem:
1. **Telemetria (Sensores):** Leituras contínuas (ex: temperatura) geradas a cada milissegundo. A velocidade da rede é crucial.
2. **Controle (Atuadores):** Comandos esporádicos (ex: "desligar caldeira"). São ações críticas que não podem ser perdidas ou corrompidas na rede sob nenhuma hipótese.

</details>

---

<details>
  <summary><h2> Arquitetura e Decisões de Design</h2></summary>

Por questões comerciais do projeto, não foi utilizado nenhum framework (como MQTT). Toda a comunicação e roteamento foram implementados através da arquitetura nativa da internet (Sockets).

- **Sensores (Telemetria via UDP - Porta 7000):** Utilizam o protocolo UDP. Como geram um imenso volume de dados a cada milissegundo, o UDP garante a velocidade necessária de entrega, aceitando perdas ocasionais sem travar o dispositivo de hardware.
- **Atuadores (Controle via TCP - Porta 9000):** Utilizam o protocolo TCP. Como comandos de controle são ações críticas, o TCP estabelece uma conexão confiável que garante a integridade e a confirmação de entrega da mensagem.
- **Aplicações Cliente (TCP - Porta 8000):** Utilizam TCP para garantir uma comunicação estável e bidirecional com o servidor, permitindo listar dados em tempo real e enviar comandos de controle.
- **Servidor de Integração (Middleware):** Centraliza as comunicações utilizando *Goroutines* para lidar com múltiplos clientes ao mesmo tempo e *Mutexes* para garantir a segurança no acesso e escrita da memória.

</details>

---

<details>
  <summary><h2> Cenário</h2></summary>

Para ilustrar a aplicação prática do middleware, o sistema desenvolvido simula o ecossistema de um **galpão industrial inteligente**. Neste ambiente, o armazenamento de produtos e a segurança do local exigem monitoramento contínuo e respostas automatizadas sem travamentos na rede.

O espaço é equipado com cinco frentes de atuação integradas pelo nosso servidor central:

1. **Sensor de Luminosidade** monitora a luz natural. Ao escurecer, acende automaticamente a **Lâmpada** do galpão.
2. **Sensor de Umidade** evita que o ar fique seco demais, acionando o **Umidificador** para proteger materiais sensíveis.
3. **Sensor de Temperatura** atua em conjunto com o **Ar Condicionado** para resfriar o ambiente e evitar o superaquecimento de máquinas e mercadorias.
4. **Sensor de Fumaça** atua como vigilante de segurança, disparando os **Sprinklers** (chuveiros de teto) de forma imediata ao detectar princípios de incêndio.
5. **Sensor de Gás** detecta vazamentos tóxicos ou inflamáveis, ligando rapidamente o **Exaustor** para sugar o ar contaminado e ventilar o prédio.

Neste cenário, o Servidor (Middleware) atua como o **cérebro do galpão**. Ele processa o alto volume de dados dos sensores e aciona os equipamentos físicos em frações de segundo, de forma totalmente autônoma. Simultaneamente, o cliente da operação pode acompanhar o status de todo o galpão e assumir o controle manual através da **Aplicação Cliente (CLI)**.

</details>

---

<details>
  <summary><h2> Funcionalidades e Automação</h2></summary>

- **Dispositivos Virtuais Simulados:** Sensores e atuadores rodam como processos em contêineres independentes, emulando perfeitamente o comportamento de um hardware real na rede.
- **Monitoramento em Tempo Real:** O cliente possui uma CLI (Interface de Linha de Comando) interativa que permite ao usuário listar e verificar todos os dispositivos ou monitorar um em específico.
- **Controle Automático:** O servidor monitora constantemente os valores recebidos e liga/desliga atuadores compatíveis automaticamente caso os limites sejam ultrapassados.
- **Controle Manual Temporário:** O usuário pode assumir o controle e enviar comandos diretos para um atuador. Quando isso ocorre, o servidor bloqueia a automação daquele atuador temporariamente para respeitar a decisão manual do usuário.

### Regras de Automação Implementadas

| Sensor | Atuador Compatível | Condição para Ligar (ON) | Condição para Desligar (OFF) |
| :--- | :--- | :--- | :--- |
| **Luminosidade** | Lâmpada | Valor < 200 lux | Valor > 300 lux |
| **Umidade** | Umidificador | Valor < 45 % | Valor > 55 % |
| **Temperatura** | Ar Condicionado | Valor > 25 °C | Valor < 20 °C |
| **Fumaça** | Sprinkler | Valor > 150 ppm | Valor < 80 ppm |
| **Gás** | Exaustor | Valor > 300 ppm | Valor < 150 ppm |

</details>

---

<details>
  <summary><h2> Guia de Uso: Executando com Docker</h2></summary>

Atendendo às restrições do projeto, o sistema foi projetado para rodar em contêineres Docker, permitindo a execução de múltiplas instâncias isoladas no laboratório de forma fácil e padronizada. O projeto já conta com um `docker-compose.yml` pré-configurado.

### 1. Construindo as imagens

Você pode construir todas as imagens do sistema de uma só vez utilizando:
```bash
docker compose build
```

Ou, se preferir, pode compilar de forma individual cada componente:
```bash
# Core
docker compose build server               # Servidor
docker compose build client               # Cliente CLI

# Sensores
docker compose build gas                  # Gás
docker compose build humidity             # Umidade
docker compose build luminosity           # Luminosidade
docker compose build smoke                # Fumaça
docker compose build temperature          # Temperatura

# Atuadores
docker compose build air_conditioner      # Ar Condicionado
docker compose build exhaust_fan          # Exaustor
docker compose build humidifier           # Umidificador
docker compose build light                # Lâmpada
docker compose build sprinkler            # Sprinkler
```

### 2. Executando o Ecossistema

**Iniciar o Servidor:**
```bash
docker compose up server
```

**Iniciar os Sensores (Terminal Interativo):**
```bash
docker compose run --rm <nome_do_sensor> ./sensor_bin <IP do servidor>
```

**Iniciar os Atuadores (Terminal Interativo):**
```bash
docker compose run --rm <nome_do_atuador> ./actuator_bin <IP do servidor>
```

**Iniciar a Aplicação Cliente (Terminal Interativo):**
```bash
docker compose run --rm client ./client_bin <IP do servidor>
```

</details>

---

<details>
  <summary><h2> Conclusão</h2></summary>

O desenvolvimento do projeto "A Rota Das Coisas" cumpriu o desafio de construir um ecossistema IoT robusto e performático sem a dependência de frameworks de terceiros. A criação de um middleware customizado permitiu resolver o grave problema de alto acoplamento da arquitetura física, poupando a memória e o processamento dos dispositivos de hardware.

A escolha estratégica e a separação dos protocolos de rede mostraram-se fundamentais para a solução do problema: o uso de **UDP** para sensores evitou o congestionamento da rede lidando de forma eficiente com o alto volume de telemetria contínua, enquanto o **TCP** garantiu a confiabilidade total exigida pelos comandos direcionados aos atuadores.

Por fim, a integração completa com o **Docker** validou o requisito arquitetural.

</details>

---

<details>
  <summary><h2> Contribuidores</h2></summary>

[<img src="https://github.com/enejota-njs.png" width="80" height="80">](https://github.com/enejota-njs)

</details>

---

<details>
  <summary><h2> Referências</h2></summary>

**Documentação Oficial da Linguagem Go (Golang)**. Disponível em: <br>
<a href="https://go.dev/doc/" target="_blank">https://go.dev/doc/</a>

</details>