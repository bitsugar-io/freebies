import { beforeEach, describe, expect, test } from 'bun:test';
import { getStateFromPath } from '@react-navigation/core';
import { LINK_CONFIG } from './linkConfig';
import { attachLinkListener, deliverLink, firstSighting, initialLink, resetLinks } from './pendingLink';

/** The leaf route a path opens, e.g. ['Tabs', 'Teams', 'Sport'], plus its params. */
function leaf(path: string) {
  let state: any = getStateFromPath(path, LINK_CONFIG);
  const names: string[] = [];
  let params: unknown;
  while (state) {
    const route = state.routes[state.routes.length - 1];
    names.push(route.name);
    params = route.params;
    state = route.state;
  }
  return { names, params };
}

describe('link paths', () => {
  const cases: [string, string[], unknown][] = [
    ['won', ['Tabs', 'Won'], undefined],
    ['tonight', ['Tabs', 'Won'], undefined],
    ['deals', ['Tabs', 'Won'], undefined],
    ['today', ['Tabs', 'Today'], undefined],
    ['teams', ['Tabs', 'Teams', 'Sports'], undefined],
    ['teams/football', ['Tabs', 'Teams', 'Sport'], { sport: 'football' }],
    ['teams/hockey', ['Tabs', 'Teams', 'Sport'], { sport: 'hockey' }],
    ['you', ['Tabs', 'You'], undefined],
    ['deal/0be2de9a-c5ff-49de-aebb-ddaeed241634', ['Deal'], { id: '0be2de9a-c5ff-49de-aebb-ddaeed241634' }],
    ['pause', ['Pause'], undefined],
  ];
  for (const [path, names, params] of cases) {
    test(path, () => {
      const got = leaf(path);
      expect(got.names).toEqual(names);
      expect(got.params).toEqual(params);
    });
  }

  test('deal links sit on top of the tabs', () => {
    const state: any = getStateFromPath('deal/abc', LINK_CONFIG);
    expect(state.routes.map((r: any) => r.name)).toEqual(['Tabs', 'Deal']);
  });
});

describe('pending links', () => {
  beforeEach(resetLinks);

  test('a link before the container mounts is used as the initial URL, once', async () => {
    deliverLink('freebie://deal/a');
    expect(await initialLink(async () => 'freebie://deal/old')).toBe('freebie://deal/a');
    expect(await initialLink(async () => 'freebie://deal/old')).toBeNull();
  });

  test('the launch URL is read once, so a remount does not replay it', async () => {
    expect(await initialLink(async () => 'freebie://deal/launch')).toBe('freebie://deal/launch');
    expect(await initialLink(async () => 'freebie://deal/launch')).toBeNull();
  });

  test('a link that lands while the launch URL is read wins', async () => {
    const got = await initialLink(async () => {
      deliverLink('freebie://teams/hockey');
      return 'freebie://deal/launch';
    });
    expect(got).toBe('freebie://teams/hockey');
  });

  test('links go straight to a mounted container; a late listener gets the held one', () => {
    const seen: string[] = [];
    deliverLink('freebie://you');
    const off = attachLinkListener((u) => seen.push(u));
    expect(seen).toEqual(['freebie://you']);
    deliverLink('freebie://won');
    expect(seen).toEqual(['freebie://you', 'freebie://won']);
    off();
    deliverLink('freebie://teams');
    expect(seen).toHaveLength(2);
    const again: string[] = [];
    attachLinkListener((u) => again.push(u));
    expect(again).toEqual(['freebie://teams']);
  });

  test('a notification is counted once', () => {
    expect(firstSighting('n1')).toBe(true);
    expect(firstSighting('n1')).toBe(false);
    expect(firstSighting(undefined)).toBe(true);
  });
});
