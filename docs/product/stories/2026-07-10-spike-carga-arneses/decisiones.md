# Decisiones — Spike carga de arneses/proyectos (2026-07-10)

> `tipo: spike`. Cada decisión conversada se escribe EN EL MISMO TURNO (§10). Cierre = decisión documentada.
> Nada firmado aún.

## S-D0 · Alcance del spike — DECIDIDA
- Investigar + **decidir el diseño** del mecanismo de carga (cara A arnés-único + cara B proyecto-multi).
- Entregar un **thin proof** de la cara A (dialog FE → PUT existente) para dar agencia inmediata.
- La cara B (detección proyecto-con-arneses) se **decide en papel**; su build es slice posterior.

## S-D1 · Estado hallado (no reconstruir) — HALLAZGO
- Backend de carga single-arnés **HECHO** (PUT valida+registra+indexa, honesto; loader CAP-15-20; FE client).
- Gaps: (1) **dialog FE** (diferido HS-07) · (2) **detector proyecto-multi** lock `.devstudio/arneses.yaml` (diferido HS-12).
- ⇒ el spike NO toca el motor de carga; ataca la afordancia FE + decide la detección de proyecto.

## S-D2 · Path input vs folder-picker nativo — ABIERTA (decidir en el proof)
- Browser (dev) no tiene folder-picker seguro; el shell de escritorio (Tauri) sí (dialog nativo).
- Hipótesis: el thin proof usa **input de path de texto** (funciona en browser Y desktop) + más tarde el
  picker nativo de Tauri como azúcar. Confirmar al construir el dialog.

## S-D3 · Cómo se deriva el `id` del arnés al cargar — ABIERTA
- El PUT exige `{id}` en la ruta. ¿El usuario lo teclea, o se **deriva** del manifiesto/nombre de carpeta
  (nomenclatura v1: manifiesto→plugin.json→id, HS-12)? Preferencia: derivar + permitir override. Decidir en el proof.

## S-D4 · Detección proyecto-con-arneses (cara B) — ABIERTA (decisión de papel)
- Un proyecto real (Vitalia) = repo con arnés(es) instalados en `.claude/` (+ posible lock `.devstudio/arneses.yaml`).
- Cargar el proyecto = detectar TODOS los arneses adentro + registrar cada uno. Requiere el detector 3° (HS-12).
- Salida esperada del spike: decidir si (i) se reusa el lock DevStudio, (ii) se camina `.claude/plugins`, o (iii) ambos;
  y si «cargar proyecto» es un endpoint nuevo o el PUT actual iterado. → documenta_decision al cerrar.

## Abiertas (siguiente)
- Construir thin proof cara A: dialog FE «Cargar carpeta» → `PUT /api/arneses/{id}` → picker+Mapa → sesión → chat.
- Con el proof en mano, cerrar S-D2/S-D3 y redactar la decisión de S-D4 (cara B).
