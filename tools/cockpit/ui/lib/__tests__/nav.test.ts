import { describe, it, expect } from 'vitest';
import { applyNavConfig, resolveSidebarNav } from '@/lib/nav';

const sistema = [{ id: 'roadmap' }, { id: 'board' }, { id: 'map' }, { id: 'evolucion' }];
const core = [{ id: 'negocio' }, { id: 'harness' }];

describe('applyNavConfig', () => {
  it('null = sin config → set completo (referencia intacta)', () => {
    expect(applyNavConfig(sistema, null)).toBe(sistema);
  });

  it('filtra y ORDENA según la lista', () => {
    expect(applyNavConfig(sistema, ['board', 'roadmap']).map((i) => i.id)).toEqual([
      'board',
      'roadmap',
    ]);
  });

  it('ignora ids desconocidos (forward-compat)', () => {
    expect(applyNavConfig(sistema, ['board', 'futuro', 'map']).map((i) => i.id)).toEqual([
      'board',
      'map',
    ]);
  });
});

describe('resolveSidebarNav (I-51 · per-board)', () => {
  it('multi: el nav per-board filtra sistema-scoped, transversales ALWAYS-ON', () => {
    const { nav, core: c } = resolveSidebarNav(sistema, core, ['board', 'evolucion'], 'multi');
    expect(nav.map((i) => i.id)).toEqual(['board', 'evolucion']);
    expect(c.map((i) => i.id)).toEqual(['negocio', 'harness']); // intacto pese al nav
  });

  it('single: el nav gobierna AMBAS familias (caso cliente)', () => {
    const { nav, core: c } = resolveSidebarNav(sistema, core, ['board', 'harness'], 'single');
    expect(nav.map((i) => i.id)).toEqual(['board']);
    expect(c.map((i) => i.id)).toEqual(['harness']); // negocio queda fuera
  });

  it('sin config (navIds null) → todo el set en cualquier modo', () => {
    const multi = resolveSidebarNav(sistema, core, null, 'multi');
    expect(multi.nav).toBe(sistema);
    expect(multi.core).toBe(core);
    const single = resolveSidebarNav(sistema, core, null, 'single');
    expect(single.core).toBe(core);
  });

  it('mode indefinido (tree cargando) → no-multi → transversales sí filtran', () => {
    const { core: c } = resolveSidebarNav(sistema, core, ['negocio'], undefined);
    expect(c.map((i) => i.id)).toEqual(['negocio']);
  });
});
