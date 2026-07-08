# PARIDAD — mockup v2 ↔ app (se llena fila por fila al implementar)

> Contrato de «lo que ves en el mockup ES lo que hace la app». Gate final: todas las
> filas ✅ + el daemon actualizándose A SÍ MISMO de verdad (binario instalado en
> ~/.local/bin, huella nueva tras el reinicio), consola limpia, screenshots revisados.

| RF | Elemento (mockup v2) | Componente real | Story/test | Estado |
|---|---|---|---|---|
| RF-100 | vista Ajustes nace con la tarjeta (223-247) | pages/shell/ui/global-view.tsx | AjustesView (view real) | 🔶 |
| RF-101 | identidad honesta: huella·ruta·repo (159-171) | update-card + GET /api/version | story Idle | 🔶 |
| RF-102 | no-escribible: disabled + migración (207-216) | update-card | story NoEscribible | 🔶 |
| RF-103 | sin repo: disabled + honesto (211) | update-card | story SinRepo | 🔶 |
| RF-104 | 5 pasos, corte al fallo, binario intacto (141-157·195-203) | ports.SelfUpdater + adapter + usecase + POST /api/self-update | go test -race ✓ + stories ErrorDeBuild/Actualizando/YaAlDia | 🔶 |
| RF-105 | reinicio + polling huella nueva (184-191) | update-card + re-exec daemon | stories Reiniciando/Exito | 🔶 |
| RF-106 | seguridad: withAuth · cero params · 409 | transport + usecase | go test (409·503·params ignorados) ✓ | 🔶 |
| RF-107 | GET /api/version (buildinfo VCS) | Go + client.ts | go test (wire RF-107) ✓ | 🔶 |

Leyenda: ⬜ pendiente · 🔶 implementado sin verificar · ✅ verificado en paridad.
