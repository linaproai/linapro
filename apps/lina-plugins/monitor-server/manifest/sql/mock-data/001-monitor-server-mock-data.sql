-- Mock data: server monitor snapshots.

INSERT IGNORE INTO plugin_monitor_server (node_name, node_ip, data, created_at, updated_at)
VALUES (
    'linapro-dev-01',
    '192.168.10.21',
    JSON_OBJECT(
        'cpu', JSON_OBJECT('usagePercent', 23.7, 'cores', 8),
        'memory', JSON_OBJECT('usedPercent', 61.4, 'totalBytes', 17179869184),
        'disk', JSON_ARRAY(JSON_OBJECT('mount', '/', 'usedPercent', 48.2, 'totalBytes', 274877906944)),
        'network', JSON_OBJECT('rxBytesPerSecond', 184320, 'txBytesPerSecond', 90112),
        'runtime', JSON_OBJECT('goroutines', 128, 'heapAllocBytes', 73400320)
    ),
    '2026-04-20 09:00:00',
    '2026-04-20 09:00:00'
);

INSERT IGNORE INTO plugin_monitor_server (node_name, node_ip, data, created_at, updated_at)
VALUES (
    'linapro-dev-02',
    '192.168.10.22',
    JSON_OBJECT(
        'cpu', JSON_OBJECT('usagePercent', 41.2, 'cores', 8),
        'memory', JSON_OBJECT('usedPercent', 72.8, 'totalBytes', 17179869184),
        'disk', JSON_ARRAY(JSON_OBJECT('mount', '/', 'usedPercent', 67.5, 'totalBytes', 274877906944)),
        'network', JSON_OBJECT('rxBytesPerSecond', 284672, 'txBytesPerSecond', 143360),
        'runtime', JSON_OBJECT('goroutines', 156, 'heapAllocBytes', 94371840)
    ),
    '2026-04-20 09:05:00',
    '2026-04-20 09:05:00'
);
