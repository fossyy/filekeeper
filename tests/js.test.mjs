// Dependency-free checks for public/main.js and public/validatePassword.js.
// Run: node --test tests/js.test.mjs
import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import path from 'node:path';
import vm from 'node:vm';
import { fileURLToPath } from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');

function load(rel, sandbox) {
    const context = vm.createContext(sandbox);
    vm.runInContext(readFileSync(path.join(root, rel), 'utf8'), context, { filename: rel });
    return context;
}

class FakeWebSocket {
    static CONNECTING = 0;
    static OPEN = 1;
    static CLOSING = 2;
    static CLOSED = 3;
    constructor(url) {
        this.url = url;
        this.readyState = FakeWebSocket.OPEN;
        this.sent = [];
    }
    send(data) {
        if (this.readyState !== FakeWebSocket.OPEN) throw new Error('not open');
        this.sent.push(data);
    }
}

function fakeElement(value = '') {
    return {
        value,
        disabled: false,
        textContent: '',
        innerText: '',
        style: {},
        classList: {
            add() {},
            remove() {},
        },
        addEventListener() {},
        dispatchEvent(event) {
            (this.dispatched ??= []).push(event);
        },
    };
}

function mainSandbox(overrides = {}) {
    return {
        console,
        WebSocket: FakeWebSocket,
        File: class File {},
        performance: { now: () => 0 },
        window: { location: { protocol: 'https:', host: 'example.test' } },
        document: { getElementById: () => null },
        Event: class Event {
            constructor(type) { this.type = type; }
        },
        FormData: class FormData {
            append() {}
        },
        fetch: async () => ({ status: 202 }),
        toggleModal: () => {},
        ...overrides,
    };
}

test('generateUniqueId returns unique uuid-v4-shaped ids', () => {
    const ctx = load('public/main.js', mainSandbox());
    const a = ctx.generateUniqueId();
    const b = ctx.generateUniqueId();
    assert.match(a, /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/);
    assert.notEqual(a, b);
});

test('convertFileSize formats bytes, KB, MB, GB', () => {
    const ctx = load('public/main.js', mainSandbox());
    assert.equal(ctx.convertFileSize(500), '500 B');
    assert.equal(ctx.convertFileSize(2048), '2.00 KB');
    assert.equal(ctx.convertFileSize(3 * 1024 * 1024), '3.00 MB');
    assert.equal(ctx.convertFileSize(2 * 1024 * 1024 * 1024), '2.00 GB');
});

test('splitFile slices full range including a short final chunk', async () => {
    const ctx = load('public/main.js', mainSandbox());
    const MB = 1024 * 1024;
    const calls = [];
    const fakeFile = {
        size: 5 * MB,
        slice(start, end) {
            calls.push([start, end]);
            return { start, end };
        },
    };
    const chunks = await ctx.splitFile(fakeFile, 2 * MB);
    assert.equal(chunks.length, 3);
    assert.deepEqual(calls, [[0, 2 * MB], [2 * MB, 4 * MB], [4 * MB, 5 * MB]]);
});

test('handleFile does not send and shows modal when socket missing', async () => {
    const modalTitle = fakeElement();
    const ctx = load('public/main.js', mainSandbox());
    ctx.document.getElementById = (id) => id === 'modalContainer'
        ? { querySelector: (selector) => (selector === '#modal h2' ? modalTitle : fakeElement()) }
        : null;
    ctx.window.mySocket = null;
    const fakeFile = new ctx.File();
    fakeFile.name = 'a.txt';
    fakeFile.size = 10;
    await ctx.handleFile(fakeFile);
    assert.equal(modalTitle.textContent, 'Upload unavailable');
    assert.ok(!ctx.window.fileIdMap || Object.keys(ctx.window.fileIdMap).length === 0);
});

test('handleFile rejects non-File input without touching the socket', async () => {
    let sent = false;
    const ctx = load('public/main.js', mainSandbox());
    ctx.window.mySocket = {
        readyState: FakeWebSocket.OPEN,
        send: () => { sent = true; },
    };
    await ctx.handleFile({ size: 10, name: 'not-a-file', slice: () => ({}) });
    assert.equal(sent, false);
});

function stubProgressAndTable(ctx, table) {
    const progress = fakeElement();
    ctx.document.getElementById = (id) => {
        if (id === 'file-table') return table;
        if (id.endsWith('-1') || id.endsWith('-2') || id.endsWith('-3')) return progress;
        return null;
    };
    return { progress, table };
}

test('uploadChunks tolerates a missing chunkArray', async () => {
    const ctx = load('public/main.js', mainSandbox());
    stubProgressAndTable(ctx, fakeElement());
    const file = { name: 'x.bin' };
    const chunks = [{ size: 10 }];
    await ctx.uploadChunks('id1', file, chunks, undefined, 'srv-id');
});

test('uploadChunks refreshes the file table via htmx.trigger when available', async () => {
    const ctx = load('public/main.js', mainSandbox());
    const triggered = [];
    ctx.window.htmx = { trigger: (el, event) => triggered.push({ el, event }) };
    const table = fakeElement();
    const { progress } = stubProgressAndTable(ctx, table);
    await ctx.uploadChunks('id2', { name: 'x.bin' }, [], {}, 'srv-id');
    assert.equal(progress.innerText, 'Done');
    assert.deepEqual(triggered, [{ el: table, event: 'load' }]);
});

test('uploadChunks falls back to dispatchEvent without htmx', async () => {
    const ctx = load('public/main.js', mainSandbox());
    const dispatched = [];
    const table = fakeElement();
    table.dispatchEvent = (event) => dispatched.push(event);
    stubProgressAndTable(ctx, table);
    await ctx.uploadChunks('id3', { name: 'x.bin' }, [], {}, 'srv-id');
    assert.equal(dispatched.length, 1);
    assert.equal(dispatched[0].type, 'load');
});

test('validatePasswords does not throw when any field is absent', () => {
    const ctx = load('public/validatePassword.js', {
        console,
        document: { getElementById: () => null },
    });
    assert.doesNotThrow(() => ctx.validatePasswords());
});

test('validatePasswords enables submit only for a valid matching password', () => {
    const elements = new Map();
    const ids = [
        'checkContainer', 'password', 'confirmPassword',
        'matchGoodPath', 'matchBadPath', 'lengthGoodPath', 'lengthBadPath',
        'uppercaseGoodPath', 'uppercaseBadPath',
        'matchSvgContainer', 'lengthSvgContainer', 'uppercaseSvgContainer',
        'numberSvgContainer', 'numberGoodPath', 'numberBadPath',
        'numberStatusText', 'matchStatusText', 'lengthStatusText', 'uppercaseStatusText', 'submit',
    ];
    for (const id of ids) elements.set(id, fakeElement());
    const ctx = load('public/validatePassword.js', {
        console,
        document: { getElementById: (id) => elements.get(id) ?? null },
    });
    const password = elements.get('password');
    const confirmPassword = elements.get('confirmPassword');
    const submit = elements.get('submit');

    password.value = 'Abcdefg1';
    confirmPassword.value = 'Abcdefg1';
    ctx.validatePasswords();
    assert.equal(submit.disabled, false);

    confirmPassword.value = 'AbcdefgX';
    ctx.validatePasswords();
    assert.equal(submit.disabled, true);

    password.value = 'Abcdefgh';
    confirmPassword.value = 'Abcdefgh';
    ctx.validatePasswords();
    assert.equal(submit.disabled, true);

    password.value = '';
    confirmPassword.value = '';
    ctx.validatePasswords();
    assert.equal(submit.disabled, true);
});
