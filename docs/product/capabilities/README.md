# capabilities/ — SSoT de "qué existe" (YAML por-cap + BDD)

> Una capacidad = un archivo `{module}/{slug}.yaml`. Schema: `../_templates/capability.template.yaml`.
> El SALDO del sistema (la story es el delta). Enforcement: `cap_doctor.py` + `docs/architecture/fitness/capability_trace_test.go`.

- **Módulos** = `project.config.yaml:domain_modules` (arnes, caja, loader, conformance, …).
- **`status` GENERADO** (R4): `vivo` (test) · `vivo·nc` (sin check) · `parcial` · `stub`. No teclear.
- **Punteros `file#Símbolo`** (R1) resuelven a código real; todo archivo reclamado por ≥1 cap (R2).
- Migración desde `CAPABILITIES.md` (82 caps) en curso — paquete `homologacion-metodologia` tarea #6.
