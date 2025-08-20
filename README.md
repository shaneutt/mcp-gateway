# MCP Gateway

An extension for [Envoy] to add support for [Model Context Protocol (MCP)][mcp].

[Envoy]:https://github.com/envoyproxy
[mcp]:https://github.com/modelcontextprotocol

## About

[Model Context Protocol (MCP)][mcp] is a standard for providing context to
Large Language Models (LLMs). This enables [Agents] to automate systems via
[MCP Servers].

In situations where there may be many MCP Servers, such as a [Kubernetes]
cluster, managing the traffic and auth for these servers is essential to
enforce limits and provide security and API management facilities for agent
traffic at scale.

This project aims to provide extensions which enable management of agent's
MCP traffic using the [Envoy] proxy server, so that Envoy can be used as an
"MCP Gateway".

[mcp]:https://github.com/modelcontextprotocol
[Agents]:https://en.wikipedia.org/wiki/Agentic_AI
[MCP Servers]:https://modelcontextprotocol.io/docs/learn/server-concepts
[Kubernetes]:https://github.com/kubernetes
[Envoy]:https://github.com/envoyproxy
