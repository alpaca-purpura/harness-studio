// Fixture Cobranza (mockup v2 GRAPHS.cobranza, franja-artefactos): un proceso ADMIN — el
// agnosticismo p7 en acción. Ejercita la casuística que el dogfood no tiene: factura
// EXTERNA de terceros sin plantilla (C4/D10), artefactos OPACOS pdf (C20), la cadena
// refina ↻v2 (C11/D9: validar-factura entrega la MISMA factura mejorada) y el gutter
// saturado de «pagar» (5 necesita → tope D11c «+N más»). Type-checked contra Graph.

import type { Graph } from "../model/types"

export const cobranzaProveedores = {
  arnes: {
    id: "cobranza-proveedores",
    rol: "Administración · Pago a proveedores",
    proceso: "de factura recibida a pago ejecutado y comprobado",
    empresa: "acme",
    reporta_a: "acme-finanzas",
    fases: ["recepción", "validación", "registro", "pago"],
    spine: {
      inicial: "llega",
      terminales: ["pagada"],
      estados: ["llega", "recibida", "validada", "registrada", "pagada"],
      transiciones: [
        { de: "llega", a: "recibida" },
        { de: "recibida", a: "validada" },
        { de: "validada", a: "registrada" },
        { de: "registrada", a: "pagada" },
      ],
    },
  },
  nodos: [
    {
      id: "recibir-factura",
      clase: "skill",
      nombre: "recibir y archivar",
      banda: "fase",
      fase: "recepción",
      estado: "llega -> recibida",
      contract: {
        caja: true,
        fase: "recepción",
        estado: "llega -> recibida",
        necesita: [{ art: "factura del proveedor", de: "terceros:proveedor", requerido: true }],
        entrega: [{ art: "factura.pdf", path: "facturas/factura.pdf", escritor_unico: true }],
      },
    },
    {
      id: "validar-factura",
      clase: "skill",
      nombre: "validar contra OC",
      banda: "fase",
      fase: "validación",
      estado: "recibida -> validada",
      contract: {
        caja: true,
        fase: "validación",
        estado: "recibida -> validada",
        necesita: [
          { art: "factura.pdf", de: "caja:recibir-factura", requerido: true },
          { art: "orden de compra", de: "terceros:compras", requerido: true },
          { art: "contrato del proveedor", de: "terceros:legal", requerido: false },
        ],
        // refina (D9): la salida ES la misma factura, mejorada (validada/anotada) → ↻v2.
        entrega: [
          {
            art: "factura.pdf",
            path: "facturas/factura.pdf",
            refina: "factura.pdf",
            escritor_unico: true,
          },
        ],
      },
    },
    {
      id: "registrar-asiento",
      clase: "skill",
      nombre: "registrar asiento",
      banda: "fase",
      fase: "registro",
      estado: "validada -> registrada",
      contract: {
        caja: true,
        fase: "registro",
        estado: "validada -> registrada",
        necesita: [{ art: "factura.pdf", de: "caja:validar-factura", requerido: true }],
        entrega: [
          {
            art: "asiento.md",
            path: "asientos/asiento.md",
            plantilla: "references/plantilla-asiento.md",
            escritor_unico: true,
          },
        ],
      },
    },
    {
      id: "pagar",
      clase: "skill",
      nombre: "ejecutar el pago",
      banda: "fase",
      fase: "pago",
      estado: "registrada -> pagada",
      contract: {
        caja: true,
        fase: "pago",
        estado: "registrada -> pagada",
        necesita: [
          { art: "asiento.md", de: "caja:registrar-asiento", requerido: true },
          { art: "factura.pdf", de: "caja:validar-factura", requerido: true },
          { art: "aprobación de gerencia", de: "usuario", requerido: true },
          { art: "datos bancarios del proveedor", de: "terceros:proveedor", requerido: false },
          { art: "constancia fiscal", de: "terceros:sat", requerido: false },
        ],
        entrega: [
          { art: "comprobante de pago.pdf", path: "pagos/comprobante.pdf", escritor_unico: true },
        ],
      },
    },
    { id: "std-pagos", clase: "rule", nombre: "política de pagos", banda: "base" },
  ],
  edges: [
    { de: "recibir-factura", a: "validar-factura", tipo: "invoca" },
    { de: "validar-factura", a: "registrar-asiento", tipo: "invoca" },
    { de: "registrar-asiento", a: "pagar", tipo: "invoca" },
    { de: "validar-factura", a: "std-pagos", tipo: "lee" },
    { de: "pagar", a: "std-pagos", tipo: "lee" },
  ],
} satisfies Graph
