import { useState } from 'react';
import Icon from '../components/Icon';
import * as api from '../../wailsjs/go/main/App';
import { config as models } from '../../wailsjs/go/models';
import type { AboutInfo, AppConfig, ScheduleStatus } from '../types';
import { shortPath } from '../utils';
import pouchBig from '../assets/brand/pouch-256.png';

const PRESETS: { label: string; times: string[] }[] = [
  { label: 'Once · midday', times: ['12:00'] },
  { label: 'Twice · morning & evening', times: ['08:00', '18:00'] },
  { label: 'Thrice · 8 · 1 · 7', times: ['08:00', '13:00', '19:00'] },
];

// The fields this page owns. Theme and category rules are edited elsewhere and
// are always taken from the freshest config when saving.
function editable(c: AppConfig): string {
  return JSON.stringify([
    c.times,
    c.instant,
    c.instantDelay,
    c.sources,
    c.minCluster,
    c.minAgeSeconds,
  ]);
}

export default function Scheduler({
  cfg,
  about,
  schedule,
  onSaved,
  notify,
}: {
  cfg: AppConfig;
  about: AboutInfo;
  schedule: ScheduleStatus;
  onSaved: () => void | Promise<void>;
  notify: (m: string) => void;
}) {
  const [draft, setDraft] = useState<AppConfig>({ ...cfg });
  const [newTime, setNewTime] = useState('12:00');
  const [saving, setSaving] = useState(false);

  const patch = (p: Partial<AppConfig>) => setDraft((d) => ({ ...d, ...p }));
  const dirty = editable(draft) !== editable(cfg);

  const addTime = () => {
    if (!/^([01]\d|2[0-3]):[0-5]\d$/.test(newTime)) {
      notify('Pick a valid time first');
      return;
    }
    if (!draft.times.includes(newTime)) {
      patch({ times: [...draft.times, newTime].sort() });
    }
  };

  const removeTime = (t: string) => patch({ times: draft.times.filter((x) => x !== t) });

  const restartAdmin = async () => {
    try {
      await api.RestartAsAdmin();
    } catch (e) {
      notify('Could not relaunch elevated: ' + String(e));
    }
  };

  const addSource = async () => {
    try {
      const p = (await api.PickFolder()) as unknown as string;
      if (!p) return;
      if (draft.sources.includes(p)) {
        notify('That folder is already watched');
        return;
      }
      patch({ sources: [...draft.sources, p] });
    } catch {
      /* cancelled */
    }
  };

  const removeSource = (p: string) => patch({ sources: draft.sources.filter((s) => s !== p) });

  // One save for the whole page: the backend writes the config and registers
  // the Windows tasks together, rolling both back if Task Scheduler refuses.
  const save = async () => {
    setSaving(true);
    try {
      // Merge with the latest cfg so a theme toggle made while this page was
      // open is never rolled back by saving automation.
      const payload: AppConfig = {
        ...cfg,
        times: draft.times,
        instant: draft.instant,
        instantDelay: draft.instantDelay,
        sources: draft.sources,
        minCluster: draft.minCluster,
        minAgeSeconds: draft.minAgeSeconds,
      };
      const saved = (await api.SaveAutomation(
        models.Config.createFrom(payload),
      )) as unknown as AppConfig;
      setDraft({ ...saved });
      await onSaved();
      const n = saved.times.length;
      notify(
        n === 0 && !saved.instant
          ? 'Automation off — manual mode.'
          : n === 0
            ? 'Instant mode on. Daily schedule off.'
            : `Saved · ${n} daily run${n > 1 ? 's' : ''} at ${saved.times.join(', ')}`,
      );
    } catch (e) {
      // Show what Windows actually accepted instead of a phantom schedule.
      try {
        const actual = (await api.GetConfig()) as unknown as AppConfig;
        setDraft({ ...actual });
      } catch {
        /* keep the draft so nothing typed is lost */
      }
      await onSaved();
      notify('Could not save automation: ' + String(e));
    } finally {
      setSaving(false);
    }
  };

  const registered = schedule.registered ?? [];
  const tasksMatch =
    schedule.times.length === registered.length &&
    schedule.times.every((t) => registered.includes(t));
  const watchOk = !schedule.instant || schedule.watchActive;
  const stateLabel = dirty ? 'unsaved changes' : tasksMatch && watchOk ? 'in sync' : 'needs save';
  const stateClass = stateLabel === 'in sync' ? 'pill' : 'pill warn';
  const watchNote = schedule.instant
    ? schedule.watchActive
      ? 'Watcher task active.'
      : 'Watcher task missing — save to repair.'
    : 'Watcher task off.';

  return (
    <div>
      <div className="row between" style={{ marginBottom: 14 }}>
        <div className="row">
          {schedule.isAdmin ? (
            <span className="pill">running as administrator</span>
          ) : (
            <span className="pill warn">standard user</span>
          )}
          {!schedule.isAdmin && (
            <button className="btn small" onClick={restartAdmin}>
              <Icon name="zap" size={15} /> Restart as administrator
            </button>
          )}
        </div>
        <div className="row">
          <span className={stateClass}>{stateLabel}</span>
          <button className="btn primary" onClick={save} disabled={saving || !dirty}>
            {saving ? <span className="spinner" /> : <Icon name="check" size={16} />} Save changes
          </button>
        </div>
      </div>

      <div className="card card-pad">
        <div className="row">
          <div className="icon-badge warm">
            <Icon name="clock" size={18} />
          </div>
          <div>
            <div style={{ fontWeight: 650 }}>Daily schedule</div>
            <div className="stat-hint">
              One Windows task per time. NEST wakes up, sorts, shows a toast — then quits.
            </div>
          </div>
        </div>

        <div className="preset-row">
          {PRESETS.map((p) => (
            <button key={p.label} className="preset-btn" onClick={() => patch({ times: p.times })}>
              {p.label}
            </button>
          ))}
          <button className="preset-btn" onClick={() => patch({ times: [] })}>
            Clear
          </button>
        </div>

        {draft.times.length === 0 && (
          <div className="empty" style={{ padding: '18px' }}>
            No daily times — manual mode (plus instant mode below, if on).
          </div>
        )}
        {draft.times.map((t) => (
          <div className="slot-row" key={t}>
            <Icon name="clock" size={16} />
            <span className="slot-time">{t}</span>
            <span style={{ flex: 1 }} />
            <button className="btn ghost small" onClick={() => removeTime(t)}>
              <Icon name="x" size={15} />
            </button>
          </div>
        ))}

        <div className="row" style={{ marginTop: 12 }}>
          <input
            className="input"
            type="time"
            value={newTime}
            onChange={(e) => setNewTime(e.target.value)}
            style={{ maxWidth: 150 }}
          />
          <button className="btn small" onClick={addTime}>
            <Icon name="plus" size={15} /> Add time
          </button>
        </div>
      </div>

      <div className="section-title">
        <Icon name="bell" size={17} /> Instant mode
        <span className="count">sort new downloads on arrival</span>
      </div>
      <div className="card card-pad">
        <div className="row between">
          <div className="row">
            <div className="icon-badge warm">
              <Icon name="zap" size={18} />
            </div>
            <div>
              <div style={{ fontWeight: 650 }}>Watch for new downloads</div>
              <div className="stat-hint">
                A logon task waits with zero CPU. When files land, NEST organizes them, shows
                one notification, and goes quiet again. No window, no tray icon.
                {!schedule.isAdmin && ' Needs administrator rights — restart above first.'}
              </div>
            </div>
          </div>
          <div
            className={`switch ${draft.instant ? 'on' : ''}`}
            onClick={() => patch({ instant: !draft.instant })}
          >
            <div className="knob" />
          </div>
        </div>
        <div className="row" style={{ marginTop: 14 }}>
          <div className="field" style={{ maxWidth: 220 }}>
            <label>Wait after last change (seconds)</label>
            <input
              className="input"
              type="number"
              min={10}
              max={3600}
              value={draft.instantDelay}
              onChange={(e) => patch({ instantDelay: Number(e.target.value) })}
            />
          </div>
        </div>
        <div className="stat-hint" style={{ marginTop: 14 }}>
          {watchNote}{' '}
          {schedule.times.length > 0
            ? `Runs daily at ${schedule.times.join(', ')}.`
            : 'No daily times.'}
        </div>
      </div>

      <div className="section-title">
        <Icon name="folder" size={17} /> Watched folders
        <span className="count">Downloads by default</span>
      </div>
      <div className="card card-pad">
        {draft.sources.length === 0 && (
          <div className="empty">No folders watched — NEST will have nothing to sort.</div>
        )}
        {draft.sources.map((s) => (
          <div className="folder-chip" key={s} style={{ marginBottom: 8 }}>
            <div className="icon-badge">
              <Icon name="folder" size={18} />
            </div>
            <div className="meta" style={{ flex: 1 }}>
              <div className="name">{shortPath(s)}</div>
              <div className="path">{s}</div>
            </div>
            <button className="btn ghost small" onClick={() => removeSource(s)}>
              <Icon name="x" size={15} />
            </button>
          </div>
        ))}
        <button className="btn small" onClick={addSource} style={{ marginTop: 8 }}>
          <Icon name="plus" size={15} /> Add folder
        </button>
      </div>

      <div className="section-title">
        <Icon name="sparkles" size={17} /> Smart rules
      </div>
      <div className="card card-pad grid grid-2">
        <div className="field">
          <label>Min files to create a cluster folder</label>
          <input
            className="input"
            type="number"
            min={1}
            value={draft.minCluster}
            onChange={(e) => patch({ minCluster: Number(e.target.value) })}
          />
        </div>
        <div className="field">
          <label>Skip files newer than (seconds)</label>
          <input
            className="input"
            type="number"
            min={0}
            max={86400}
            value={draft.minAgeSeconds}
            onChange={(e) => patch({ minAgeSeconds: Number(e.target.value) })}
          />
        </div>
      </div>

      <div className="row" style={{ marginTop: 16 }}>
        <button className="btn" onClick={save} disabled={saving || !dirty}>
          {saving ? <span className="spinner" /> : <Icon name="check" size={16} />} Save changes
        </button>
        <span className="stat-hint">
          Saves times, instant mode, folders, delay and smart rules together.
        </span>
      </div>

      <div className="section-title">
        <Icon name="zap" size={17} /> About
      </div>
      <div className="card card-pad">
        <div className="about-grid">
          <img src={pouchBig} alt="NEST" />
          <div>
            <div style={{ fontWeight: 700, fontSize: 16 }}>
              {about.name} <span className="stat-hint">v{about.version}</span>
            </div>
            <div className="stat-hint" style={{ marginTop: 4 }}>
              Designed and built by <strong style={{ color: 'var(--orange)' }}>{about.author}</strong>
            </div>
            <div className="mono" style={{ marginTop: 8 }}>
              {about.exePath}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
