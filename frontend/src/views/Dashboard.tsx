import Icon from '../components/Icon';
import type { DashboardData, RunRecord } from '../types';
import { formatBytes, timeAgo, shortPath } from '../utils';

export default function Dashboard({
  data,
  history,
  onOpen,
}: {
  data: DashboardData | null;
  history: RunRecord[];
  onOpen: (p: string) => void;
}) {
  if (!data) {
    return <div className="empty">Loading…</div>;
  }

  const byCat = new Map<string, number>();
  history.forEach((r) => {
    if (!r.undone) r.moves.forEach((m) => byCat.set(m.category, (byCat.get(m.category) ?? 0) + 1));
  });
  const top = [...byCat.entries()].sort((a, b) => b[1] - a[1]).slice(0, 5);
  const max = top.length > 0 ? top[0][1] : 1;

  const last = data.lastRun;
  const lastMoves = last?.moves?.length ?? 0;
  const times = data.schedule?.times ?? [];
  const schedValue = times.length > 0 ? `${times.length}× daily` : data.schedule?.instant ? 'instant' : 'off';
  const schedHint =
    times.length > 0
      ? times.join(' · ')
      : data.schedule?.instant
        ? 'sorts new downloads'
        : 'manual only';
  const sourceHint =
    data.sources.length > 1
      ? `${data.sources.length} folders`
      : shortPath(data.sources[0] ?? '—');

  const stats = [
    {
      icon: 'download',
      warm: false,
      label: 'In source',
      value: String(data.downloadsCount),
      hint: sourceHint,
    },
    {
      icon: 'clock',
      warm: false,
      label: 'Last run',
      value: last ? timeAgo(last.timestamp) : 'never',
      hint: last ? `${lastMoves} file(s)` : 'not organized yet',
    },
    {
      icon: 'archive',
      warm: false,
      label: 'Organized',
      value: formatBytes(data.totalBytes),
      hint: `${data.runCount} run(s) logged`,
    },
    {
      icon: 'bell',
      warm: true,
      label: 'Automation',
      value: schedValue,
      hint: schedHint,
    },
  ];

  return (
    <div>
      <div className="grid grid-4">
        {stats.map((s) => (
          <div className="card card-pad" key={s.label}>
            <div className="row between">
              <div className="stat-label">{s.label}</div>
              <div className={`icon-badge${s.warm ? ' warm' : ''}`}>
                <Icon name={s.icon} size={18} />
              </div>
            </div>
            <div className="stat-value">{s.value}</div>
            <div className="stat-hint">{s.hint}</div>
          </div>
        ))}
      </div>

      <div className="section-title">
        <Icon name="folder" size={17} /> Where your files go
        <span className="count">real Windows folders</span>
      </div>
      <div className="grid grid-3">
        {data.knownFolders.map((f) => (
          <div
            className="folder-chip"
            key={f.key}
            onClick={() => onOpen(f.path)}
            style={{ cursor: 'pointer' }}
          >
            <div className="icon-badge">
              <Icon name={f.icon} size={18} />
            </div>
            <div className="meta">
              <div className="name">{f.name}</div>
              <div className="path">{shortPath(f.path)}</div>
            </div>
          </div>
        ))}
      </div>

      {top.length > 0 && (
        <>
          <div className="section-title">
            <Icon name="layers" size={17} /> Top categories
            <span className="count">all time</span>
          </div>
          <div className="card card-pad">
            {top.map(([cat, n], i) => (
              <div className="statbar-row" key={cat}>
                <span>{cat}</span>
                <div className="statbar-track">
                  <div
                    className={`statbar-fill${i === 0 ? ' top' : ''}`}
                    style={{ width: `${Math.max(4, Math.round((n / max) * 100))}%` }}
                  />
                </div>
                <span className="statbar-count">{n}</span>
              </div>
            ))}
          </div>
        </>
      )}

      {last && lastMoves > 0 && (
        <>
          <div className="section-title">
            <Icon name="history" size={17} /> Latest run
            <span className="count">{timeAgo(last.timestamp)}</span>
          </div>
          <div className="card card-pad">
            {last.moves.slice(0, 6).map((m, i) => (
              <div className="move-line" key={i} style={{ padding: '5px 0' }}>
                <span>{shortPath(m.from)}</span>
                <span className="arrow">→</span>
                <span>{shortPath(m.to)}</span>
              </div>
            ))}
            {lastMoves > 6 && (
              <div className="stat-hint" style={{ marginTop: 10 }}>
                +{lastMoves - 6} more…
              </div>
            )}
          </div>
        </>
      )}
    </div>
  );
}
