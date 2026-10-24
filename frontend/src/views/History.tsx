import { useState } from 'react';
import Icon from '../components/Icon';
import type { RunRecord } from '../types';
import { formatBytes, timeAgo, shortPath } from '../utils';

export default function History({
  history,
  onUndo,
  onClear,
  onReveal,
}: {
  history: RunRecord[];
  onUndo: () => void;
  onClear: () => void;
  onReveal: (p: string) => void;
}) {
  const [open, setOpen] = useState<string | null>(null);
  const [confirming, setConfirming] = useState(false);
  const lastUndoable = history.find((h) => !h.undone);
  const lastRunKey = lastUndoable ? lastUndoable.runId || lastUndoable.id : '';
  // One run may span several folders; undo reverts all of them.
  const lastCount = lastUndoable
    ? history
        .filter((h) => !h.undone && (h.runId || h.id) === lastRunKey)
        .reduce((n, h) => n + h.moves.length, 0)
    : 0;

  const askClear = () => {
    if (confirming) {
      setConfirming(false);
      onClear();
      return;
    }
    setConfirming(true);
    window.setTimeout(() => setConfirming(false), 4000);
  };

  if (history.length === 0) {
    return <div className="empty">No runs yet. NEST hasn't organized anything.</div>;
  }

  return (
    <div>
      <div className="row between" style={{ marginBottom: 16 }}>
        <div className="stat-hint">
          {lastUndoable
            ? `Last run moved ${lastCount} file(s). You can put them all back.`
            : 'Clearing wipes the log and resets the stats.'}
        </div>
        <div className="row">
          {lastUndoable && (
            <button className="btn" onClick={onUndo}>
              <Icon name="undo" size={16} /> Undo last run
            </button>
          )}
          <button className="btn ghost" onClick={askClear}>
            <Icon name="x" size={15} /> {confirming ? 'Click again to clear' : 'Clear history'}
          </button>
        </div>
      </div>

      {history.map((run) => (
        <div className="history-item" key={run.id}>
          <div className="history-head">
            <div className="icon-badge">
              <Icon name="history" size={17} />
            </div>
            <div style={{ flex: 1 }}>
              <div style={{ fontWeight: 650 }}>
                {run.moves.length} file(s) · {formatBytes(run.bytes)}
              </div>
              <div className="stat-hint">
                {timeAgo(run.timestamp)} · {shortPath(run.source)}
              </div>
            </div>
            {run.undone ? (
              <span className="pill gray">undone</span>
            ) : (
              <span className="pill">done</span>
            )}
            <button
              className="btn ghost small"
              onClick={() => setOpen(open === run.id ? null : run.id)}
            >
              {open === run.id ? 'Hide' : 'Details'}
            </button>
          </div>

          {open === run.id && (
            <div className="history-moves">
              {run.moves.map((m, i) => (
                <div className="move-line" key={i}>
                  <span>{shortPath(m.from)}</span>
                  <span className="arrow">→</span>
                  <span style={{ flex: 1 }}>{shortPath(m.to)}</span>
                  {!run.undone && (
                    <button
                      className="icon-btn"
                      title="Reveal in Explorer"
                      onClick={() => onReveal(m.to)}
                    >
                      <Icon name="folder" size={14} />
                    </button>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      ))}
    </div>
  );
}
