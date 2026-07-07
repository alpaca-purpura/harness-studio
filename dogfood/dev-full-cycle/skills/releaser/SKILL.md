---
name: releaser
nombre: promover a released
description: Publica el arnés como snapshot inmutable versionado. Usar cuando el veredicto de review es verde y la unidad de trabajo entra a la fase release.
version: 1.0
clase: skill
contract:
  why: "publicar el arnés como snapshot inmutable versionado"
  clase: skill
  arquetipo: pipeline
  perfil_harness: T1
  caja: true
  fase: release
  estado: "review -> released"
  necesita:
    - art: "veredicto de review"
      de: "caja:reviewer"
      requerido: true
  entrega:
    - art: "release@version"
      escritor_unico: true
  gate:
    tipo: none
    detalle: "gate de fidelidad (§8.4) aún no operacionalizado — diferido honesto"
---

# releaser — la caja de la fase `release` del arnés dev-full-cycle

Última caja del arnés dogfood. Solo opera con un «veredicto de review» verde de `reviewer`;
sin veredicto no hay release (precondición requerida, jamás se salta).

Arquetipo pipeline, perfil T1: pase único y determinista — corta la versión (semver), etiqueta
el snapshot inmutable y lo publica como `release@version` por el release train del kit. Cero
juicio creativo: si algo del empaquetado falla, se detiene y reporta; no repara ni re-decide.

Gate `none` declarado honesto: el gate de fidelidad (§8.4) aún no está operacionalizado — el
hueco es dato visible del contrato, no un eval fabricado.
