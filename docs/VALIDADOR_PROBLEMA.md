# Validador del problema — Hackathon QIA / Serfinanza

Documento de **validación cruzada** entre el planteamiento oficial de los retos (Partes 1, 2 y 3) y la solución implementada en este repositorio (`hackathon-bqia`).

**Última revisión:** 2026-05-23  
**Alcance del repo:** Backend **Agente 360** (Go + RAG + WhatsApp vía Evolution API)

---

## Resumen ejecutivo

| Reto | ¿Lo cubre este repo? | Nivel de cumplimiento |
|------|----------------------|------------------------|
| **Reto #3 — GEO / SEO** | No (fuera de alcance técnico) | ❌ No implementado aquí |
| **Reto Agente 360** | Sí (núcleo de la solución) | ✅ Parcial–Alto (MVP funcional) |
| **Macro-problema unificado** | Parcial (solo cara interna/omnicanal) | ⚠️ Requiere frente web GEO aparte |

**Conclusión:** Este repositorio responde de forma **demostrable** al dolor del **Agente 360** (conocimiento fragmentado, latencia, inconsistencia entre canales). El **Reto #3 (GEO)** debe validarse en otro artefacto (WordPress, Search Console, auditorías LLM externas).

---

## Leyenda de validación

| Símbolo | Significado |
|---------|-------------|
| ✅ | Cumple o está demostrado en código/despliegue |
| ⚠️ | Cumple parcialmente / requiere trabajo adicional |
| ❌ | No cumple o está fuera de alcance de este repo |
| 🔬 | Requiere prueba manual o métrica en demo |

---

# PARTE 1 — Reto #3: GEO / SEO  
*"Invisibilidad en la era de los Motores Generativos"*

## 1. Enfoque en la necesidad (el dolor)

| Criterio | Estado | Evidencia / nota |
|----------|--------|------------------|
| Usuarios piden respuestas a LLMs, no solo enlaces | ✅ (problema real) | Planteamiento válido del reto |
| Banco invisible en recomendaciones de IA | ⚠️ | No medible desde este backend |
| Web no legible para rastreadores de IA | ❌ en este repo | No hay WordPress ni auditoría SEO aquí |

## 2. Delimitación y alcance

| Dentro del reto | Estado en `hackathon-bqia` |
|-----------------|----------------------------|
| Auditar salud técnica WordPress | ❌ |
| Corregir indexación / metadatos | ❌ |
| Contenido “citable” para agentes IA | ⚠️ | El `knowledge.json` **sí** modela chunks citables, pero es corpus **interno**, no web pública |
| Rediseño gráfico / campañas pagas | ❌ (correctamente fuera) | N/A |

## 3. Medible y basado en datos

| Métrica objetivo | Línea base declarada | Estado en este repo |
|------------------|----------------------|---------------------|
| Salud técnica web > 90% | 74% actual | ❌ No aplica |
| Mitigar 1.009 advertencias | — | ❌ No aplica |
| Aparición en auditorías LLM simuladas | Desconocida/nula | ❌ No instrumentado aquí |

## 4. Restricciones

| Restricción | Cumplimiento |
|-------------|--------------|
| Sprint 48 h | 🔬 Depende del equipo |
| Google Analytics / Search Console | ❌ No integrado en este repo |
| No alterar BD de clientes / datos sensibles | ✅ | RAG usa JSON estático; sin PII transaccional |

### Veredicto Parte 1

> **Este repositorio NO valida el Reto #3.**  
> Para el jurado: presentar evidencia en otro entregable (informe SEO/GEO, capturas Search Console, pruebas de citación en ChatGPT/Gemini/Perplexity).

---

# PARTE 2 — Reto: Agente 360  
*"Fragmentación y latencia en el acceso al conocimiento institucional"*

## 1. Enfoque en la necesidad (el dolor)

| Dolor declarado | ¿Lo ataca la solución? | Cómo |
|-----------------|------------------------|------|
| Información fragmentada en silos | ✅ | Corpus unificado en `data/knowledge.json` + perfiles en `data/profiles.json` |
| Asesores tardan en encontrar respuestas | ✅ | API `/ask` con RAG + respuesta en ~1 s (Postman/VPS) |
| Respuestas inconsistentes por canal | ⚠️ | Misma fuente RAG para API y WhatsApp; falta integrar oficina/App formalmente |
| Onboarding lento de empleados | ⚠️ | Demostrable conceptualmente; sin métricas de adopción |

## 2. Delimitación y alcance

| Dentro | Estado | Implementación |
|--------|--------|----------------|
| Arquitectura de consulta centralizada | ✅ | `internal/rag/retrieve.go`, `internal/handlers/handlers.go` |
| RAG sobre documentación del banco | ✅ | Chunks con `doc`, `seccion`, `tags`, `contenido` |
| Omnicanal (empleados + clientes) | ⚠️ | Canales: REST (`/ask`) + WhatsApp (`/whatsapp/webhook`); falta UI Angular/Copilot |
| Proactivo según perfil | ✅ | `ProactiveHint`, `/recommend`, recomendaciones por producto |
| Modificar core transaccional | ❌ (correcto) | No hay integración core |

| Fuera | Respetado |
|-------|-----------|
| Core de créditos / desembolsos | ✅ |

## 3. Medible y basado en datos

| KPI declarado | Estado | Nota para demo |
|---------------|--------|----------------|
| Reducción tiempo de búsqueda asesor | 🔬 | Medir antes/después en demo con cronómetro |
| Resolución en primer contacto | 🔬 | Simular 5 preguntas frecuentes vía `/ask` |
| Contradicciones entre canales → ~0% | ⚠️ | Mismo RAG; validar que WhatsApp y `/ask` den misma respuesta a misma pregunta |
| Menor tiempo onboarding | 🔬 | Narrativa + guía de uso |

## 4. Restricciones y fricciones

| Restricción institucional | Cumplimiento en este repo |
|---------------------------|---------------------------|
| Stack Angular + .NET + Copilot Studio | ❌ | Stack real: **Go 1.22**, Chi, Anthropic Claude |
| Interfaz intuitiva para personal | ⚠️ | API lista; UI no está en este repo |
| Resistencia al cambio | 🔬 | WhatsApp reduce fricción para cliente piloto |

### Veredicto Parte 2

> **El reto Agente 360 está validado a nivel MVP backend.**  
> Gap principal: stack exigido (.NET/Angular) y métricas formales aún no instrumentadas.

---

# PARTE 3 — Unificación de los dos problemas  
*"La brecha de accesibilidad al conocimiento financiero"*

## Mapa conceptual vs. implementación

```
[ INFORMACIÓN DEL BANCO ]
           |
    +------+------+
    |             |
 EXTERNO       INTERNO
 (GEO/SEO)   (Agente 360)
    |             |
    ❌            ✅ ← este repositorio
 WordPress     Go + RAG + WhatsApp
```

| Dimensión | Objetivo unificado | Este repo |
|-----------|-------------------|-----------|
| **Atracción (externo)** | Salud web >90%, menciones en LLMs | ❌ |
| **Operación (interno)** | Cero contradicciones, menos tiempo atención | ⚠️–✅ |
| **Estrategia simbiótica** | GEO público + RAG privado | Solo mitad (RAG) |

## Veredicto Parte 3

La narrativa unificada es **coherente** para el pitch, pero este código **solo demuestra la cara interna/omnicanal del Agente 360**. El frente GEO debe mostrarse como trabajo complementario del equipo.

---

# Validación técnica de la solución implementada

## Arquitectura actual

| Componente | Archivo / ruta | Función |
|------------|----------------|---------|
| API HTTP | `internal/server/router.go` | Endpoints REST |
| RAG | `internal/rag/retrieve.go` | Recuperación por score + sinónimos |
| Alcance / guardrails | `internal/rag/scope.go` | Fuera de tema, saludos, score mínimo |
| LLM | `internal/llm/anthropic.go` | Claude + formato WhatsApp |
| Sesión | `internal/session/store.go` | Historial por `sessionId` / teléfono |
| WhatsApp | `internal/whatsapp/webhook.go` | Parse Evolution + filtro por número |
| Corpus | `data/knowledge.json` | ~114 líneas de conocimiento institucional |
| Perfiles | `data/profiles.json` | Clientes C001–C004 |
| Deploy | `docker-compose.yml`, `commands/deploy.sh` | VPS puerto **8090** |

## Endpoints y criterios de aceptación

| Endpoint | Método | Criterio de éxito | 🔬 Prueba |
|----------|--------|-------------------|----------|
| `/health` | GET | `200` + `status: ok` | `curl http://144.91.79.105:8090/health` |
| `/ask` | POST | Respuesta con `answer`, `citations`, `grounded` | Postman collection |
| `/simulate-cdt` | POST | Simulación CDT | Body con monto/plazo |
| `/recommend` | POST | Recomendación por `profileId` | `C004` |
| `/whatsapp/webhook` | GET | `webhook_verified` | `curl` GET |
| `/whatsapp/webhook` | POST | `received: true` + `answer` | JSON Evolution o simple |
| `/whatsapp/webhook/573168731521` | POST | Mismo + filtro URL | Número autorizado |

## Checklist de calidad RAG (Agente 360)

| # | Criterio | Estado | Evidencia |
|---|----------|--------|-----------|
| 1 | Respuestas citan documento y sección | ✅ | Campo `citations` en `/ask` |
| 2 | No inventa fuera del corpus (`grounded`) | ✅ | `rag.IsInScope` + umbral score 3.0 |
| 3 | Rechaza temas off-topic | ✅ | `OutOfScopeReply` en `scope.go` |
| 4 | Personalización por perfil | ✅ | `profileId`, `ProactiveHint` |
| 5 | Canal WhatsApp unificado | ⚠️ | Backend OK; Evolution debe tener webhook ON + `MESSAGES_UPSERT` |
| 6 | Un solo número piloto (seguridad) | ✅ | `WHATSAPP_ALLOWED_NUMBER=573168731521` |
| 7 | Datos en contenedor Docker | ✅ | `COPY data/` en Dockerfile (post-fix) |

## Checklist de despliegue (VPS)

| # | Verificación | Comando / acción |
|---|--------------|------------------|
| 1 | Contenedor arriba | `docker ps \| grep hackathon-qia` |
| 2 | Puerto 8090 libre y expuesto | `curl :8090/health` |
| 3 | Red `banco-agent-net` | `docker network inspect banco-agent-net` |
| 4 | Sin error `profiles.json` | Logs sin `no such file or directory` |
| 5 | Webhook Evolution activo | Enabled ON + `MESSAGES_UPSERT` ON |
| 6 | POST en logs al escribir WhatsApp | `docker compose logs -f` → línea `whatsapp webhook: mensaje de` |

## Matriz de diagnóstico WhatsApp (problema operativo resuelto)

| Síntoma en logs | Causa probable | Acción |
|-----------------|----------------|--------|
| Solo `GET /health` | Evolution no envía eventos | Activar webhook + `MESSAGES_UPSERT` |
| `GET/HEAD /whatsapp/webhook` → 405 | Verificación sin ruta GET | Actualizar backend (ya corregido) |
| `action: ignored` sin `reason` | Body JSON inválido en prueba | Usar payload Evolution en Postman |
| `reason: unauthorized_number` | Número distinto al permitido | Escribir desde `573168731521` |
| `whatsapp send error` | Fallo Evolution al responder | Revisar instancia `javierg`, apikey, puerto 8083 |
| Postman OK, WhatsApp no | Webhook Evolution desactivado | Panel Evolution → Save |

---

# Brechas para cerrar antes de la defensa

| Prioridad | Brecha | Acción sugerida |
|-----------|--------|-----------------|
| Alta | Reto GEO sin artefacto | Entregar informe SEO/GEO + capturas Search Console |
| Alta | Stack .NET/Angular no usado | Justificar PoC en Go o mostrar frontend Angular separado |
| Media | Métricas KPI sin baseline | 3 pruebas cronometradas en demo (`/ask` vs. búsqueda manual) |
| Media | WhatsApp real | Confirmar Evolution webhook + log `POST` |
| Baja | Más chunks en `knowledge.json` | Ampliar corpus para más productos |
| Baja | `ANTHROPIC_API_KEY` en `.env.docker` | Key real en VPS (no placeholder) |

---

# Guión de validación en vivo (5 minutos)

1. **Problema (30 s):** “El conocimiento del banco está fragmentado; los asesores y clientes tardan y contradicen respuestas.”
2. **Solución (30 s):** “Agente 360: un cerebro RAG central con la misma verdad en API y WhatsApp.”
3. **Demo `/ask` (1 min):** Pregunta “¿Qué es un CDT?” → mostrar `answer`, `citations`, `grounded: true`.
4. **Demo perfil (1 min):** `/recommend` con `C004` → recomendación proactiva.
5. **Demo WhatsApp (1 min):** Mensaje desde `573168731521` → log + respuesta en teléfono.
6. **Cierre unificado (1 min):** “Este repo cubre operación interna; el frente GEO abre la puerta externa en WordPress (>90% salud).”

---

# Declaración de validez del problema

| Afirmación del planteamiento | ¿Es válida? | ¿La solución la aborda? |
|------------------------------|-------------|-------------------------|
| El consumo de información migra a LLMs | ✅ Sí | Parcial (Agente propio, no GEO público) |
| El banco es invisible para IA externas | ✅ Plausible | ❌ No en este repo |
| El conocimiento interno está fragmentado | ✅ Sí | ✅ Corpus unificado |
| Los canales dan respuestas distintas | ✅ Sí | ⚠️ Unificados en RAG, falta más canales |
| Hay restricción de 48 h y seguridad | ✅ Sí | ✅ Sin tocar core ni PII transaccional |
| Stack obligatorio .NET/Angular | ✅ Como requisito | ❌ Desviación técnica a documentar |

---

## Referencias del repositorio

- README: endpoints y stack
- Postman: `hackaton_collection.json`
- Deploy: `./commands/deploy.sh up`
- Variables: `.env.example`, `.env.docker`

---

*Documento generado para alinear el pitch del hackathon con evidencia técnica verificable. Actualizar tras cada hito de demo (WhatsApp en producción, métricas, frente GEO).*
