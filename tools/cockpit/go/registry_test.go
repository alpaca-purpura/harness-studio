package main

import (
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// CK-03/CK-05: registryProject declara los campos directorio en el named field
// `Directorio` con yaml:",inline" — el wire format de ~/.cockpit/cockpit.yaml se
// mantiene FLAT (kind/sistemas/repo al nivel del proyecto). Este test es la red
// del inline (CK-03 lo verificó en sesión pero nunca se commiteó): cockpit.yaml
// viejo (sin campos directorio) y nuevo (flat) parsean, y Marshal jamás anida
// un bloque `directorio:`.
func TestRegistryWireFormatRoundTrip(t *testing.T) {
	t.Run("viejo sin campos directorio → zero-value", func(t *testing.T) {
		raw := "projects:\n- name: demo\n  path: /tmp/demo\n  active: true\n"
		var reg registry
		if err := yaml.Unmarshal([]byte(raw), &reg); err != nil {
			t.Fatal(err)
		}
		if len(reg.Projects) != 1 {
			t.Fatalf("esperaba 1 proyecto, obtuve %d", len(reg.Projects))
		}
		p := reg.Projects[0]
		if p.Name != "demo" || p.Path != "/tmp/demo" || !p.Active {
			t.Errorf("genéricos mal parseados: %+v", p)
		}
		if !reflect.DeepEqual(p.Directorio, directorioMeta{}) {
			t.Errorf("sin campos directorio debe quedar zero-value: %+v", p.Directorio)
		}
	})

	t.Run("nuevo flat → campos caen en Directorio", func(t *testing.T) {
		raw := "projects:\n" +
			"- name: prenter\n  path: /tmp/prenter\n  active: true\n" +
			"  nombre: Prenter · la EMPRESA\n  gate_check: python3 harnesses/scripts/check.py\n" +
			"  kind: factory\n  repo: /tmp/prenter-shell\n" +
			"  sistemas:\n" +
			"  - slug: odoo\n    procedencia: compartido\n    ref: odoo#company=Prenter\n" +
			"servicios_compartidos:\n" +
			"- slug: odoo\n  nombre: Odoo\n  consumido_por: [prenter]\n"
		var reg registry
		if err := yaml.Unmarshal([]byte(raw), &reg); err != nil {
			t.Fatal(err)
		}
		p := reg.Projects[0]
		if p.Directorio.Kind != "factory" || p.Directorio.Repo != "/tmp/prenter-shell" {
			t.Errorf("kind/repo flat deben caer en Directorio: %+v", p.Directorio)
		}
		if len(p.Directorio.Sistemas) != 1 || p.Directorio.Sistemas[0].Slug != "odoo" {
			t.Errorf("sistemas flat mal parseados: %+v", p.Directorio.Sistemas)
		}
		if len(reg.ServiciosCompartidos) != 1 || reg.ServiciosCompartidos[0].Nombre != "Odoo" {
			t.Errorf("servicios_compartidos mal parseados: %+v", reg.ServiciosCompartidos)
		}
		// Aditivos de la torre (RN-06): nombre + gate_check caen en el proyecto, no en Directorio.
		if p.Nombre != "Prenter · la EMPRESA" || p.GateCheck != "python3 harnesses/scripts/check.py" {
			t.Errorf("nombre/gate_check aditivos mal parseados: %+v", p)
		}
	})

	t.Run("Marshal preserva flat y round-trippea", func(t *testing.T) {
		orig := registry{
			Projects: []registryProject{{
				Name: "prenter", Path: "/tmp/prenter", Active: true,
				Nombre: "Prenter · la EMPRESA", GateCheck: "python3 harnesses/scripts/check.py",
				Directorio: directorioMeta{
					Kind: "factory", Repo: "/tmp/prenter-shell",
					Sistemas: []sistemaMeta{{
						Slug: "prenter-harness", Procedencia: "propio", Workspace: "/tmp/ph",
					}},
				},
			}},
			ServiciosCompartidos: []servicioCompartido{{Slug: "odoo", ConsumidoPor: []string{"prenter"}}},
		}
		out, err := yaml.Marshal(&orig)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(out), "directorio:") {
			t.Fatalf("el inline se rompió — apareció un bloque anidado:\n%s", out)
		}
		if !strings.Contains(string(out), "kind: factory") {
			t.Fatalf("kind debe serializar flat al nivel del proyecto:\n%s", out)
		}
		if !strings.Contains(string(out), "gate_check: python3 harnesses/scripts/check.py") {
			t.Fatalf("gate_check debe serializar flat al nivel del proyecto (RN-06):\n%s", out)
		}
		var back registry
		if err := yaml.Unmarshal(out, &back); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(orig, back) {
			t.Errorf("round-trip no idéntico:\norig: %+v\nback: %+v", orig, back)
		}
	})
}
