import { useCallback, useEffect, useRef, useState } from 'react';
import Icon from './components/Icon';
import Dashboard from './views/Dashboard';
import Categories from './views/Categories';
import Scheduler from './views/Scheduler';
import History from './views/History';
import * as api from '../wailsjs/go/main/App';
import { config as models } from '../wailsjs/go/models';
import type { AboutInfo, AppConfig, DashboardData, Report, RunRecord, UndoResult } from './types';
import { formatBytes, baseName, shortPath } from './utils';
import pouch from './assets/brand/pouch-128.png';
import pouchBig from './assets/brand/pouch-256.png';
import './style.css';

type Page = 'dashboard' | 'categories' | 'scheduler' | 'history';

const NAV: { id: Page; label: string; icon: string }[] = [
  { id: 'dashboard', label: 'Dashboard', icon: 'grid' },
  { id: 'categories', label: 'Categories', icon: 'layers' },
  { id: 'scheduler', label: 'Automation', icon: 'clock' },
  { id: 'history', label: 'History', icon: 'history' },
];

const TITLES: Record<Page, { title: string; sub: string }> = {
  dashboard: { title: 'Dashboard', sub: 'Your files, neatly back where they belong.' },
  categories: { title: 'Categories', sub: 'Extension families and where each one lands.' },
  scheduler: { title: 'Automation', sub: 'Scheduled times plus instant download sorting.' },
  history: { title: 'History', sub: 'Every move is logged and reversible.' },
};

export default function App() {
  const [page, setPage] = useState<Page>('dashboard');
  const [data, setData] = useState<DashboardData | null>(null);
  const [cfg, setCfg] = useState<AppConfig | null>(null);
  const [about, setAbout] = useState<AboutInfo | null>(null);
  const [history, setHistory] = useState<RunRecord[]>([]);
  const [busy, setBusy] = useState(false);
  const [toast, setToast] = useState('');
  const [report, setReport] = useState<Report[] | null>(null);
  const [welcome, setWelcome] = useState(false);
  const toastTimer = useRef<number | undefined>(undefined);

  const notify = useCallback((msg: string) => {
    setToast(msg);
    // A newer message must not be erased by the previous message's timer.
    if (toastTimer.current) window.clearTimeout(toastTimer.current);
    toastTimer.current = window.setTimeout(() => setToast(''), 3400);
  }, []);

  useEffect(() => () => {
    if (toastTimer.current) window.clearTimeout(toastTimer.current);
  }, []);

  const refresh = useCallback(async () => {
    try {
      const [d, c, h, ab] = await Promise.all([
        api.GetDashboard(),
        api.GetConfig(),
        api.GetHistory(),
        api.GetAbout(),
      ]);
      setData(d as unknown as DashboardData);
      setCfg(c as unknown as AppConfig);
      setHistory((h as unknown as RunRecord[]) ?? []);
      setAbout(ab as unknown as AboutInfo);
      const runs = (h as unknown as RunRecord[]) ?? [];
      if (runs.length === 0 && !window.localStorage.getItem('nest-welcomed')) {
        setWelcome(true);
      }
    } catch (e) {
      notify('Could not load NEST data: ' + String(e));
    }
  }, [notify]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', cfg?.theme === 'light' ? 'light' : 'dark');
  }, [cfg?.theme]);

  const toggleTheme = async () => {
    if (!cfg) return;
    const next = { ...cfg, theme: cfg.theme === 'light' ? 'dark' : 'light' };
    setCfg(next);
    try {
      await api.SaveConfig(models.Config.createFrom(next));
    } catch (e) {
      notify('Could not save theme: ' + String(e));
    }
  };

  const dismissWelcome = () => {
    window.localStorage.setItem('nest-welcomed', '1');
    setWelcome(false);
  };

  const onPreview = async () => {
    setBusy(true);
    try {
      const r = (await api.PreviewOrganize()) as unknown as Report[];
      setReport(r);
      const count = r.reduce((n, x) => n + x.moves.filter((m) => !m.err).length, 0);
      notify(count ? `${count} file(s) ready to organize` : 'Everything is already in place');
    } catch (e) {
      notify('Preview failed: ' + String(e));
    } finally {
      setBusy(false);
    }
  };

  const onRun = async () => {
    setBusy(true);
    try {
      const r = (await api.RunOrganize()) as unknown as Report[];
      const moved = r.reduce((n, x) => n + x.moved, 0);
      const young = r.reduce((n, x) => n + (x.skippedYoung ?? 0), 0);
      const errors = r.flatMap((x) => x.errors ?? []);
      const failed = errors.length ? ` · ${errors.length} could not be moved` : '';
      notify(
        moved
          ? `Organized ${moved} file(s)${failed}. NEST is now idle.`
          : young
            ? `${young} file(s) still downloading — try again shortly.`
            : errors.length
              ? `Could not move ${errors.length} file(s): ${errors[0]}`
              : 'Nothing to organize.',
      );
      await refresh();
    } catch (e) {
      notify('Run failed: ' + String(e));
    } finally {
      setBusy(false);
    }
  };

  const onUndo = async () => {
    try {
      const res = (await api.UndoLast()) as unknown as UndoResult;
      const gone = res.missing ? ` · ${res.missing} already gone` : '';
      notify(`Reverted ${res.moves} file(s)${gone}`);
      await refresh();
    } catch (e) {
      const msg = String(e);
      notify(msg.includes('does not exist') ? 'Nothing to undo' : 'Undo failed: ' + msg);
    }
  };

  const onClearHistory = async () => {
    try {
      await api.ClearHistory();
      notify('History cleared. Stats reset.');
      await refresh();
    } catch (e) {
      notify('Could not clear history: ' + String(e));
    }
  };

  const onReveal = async (p: string) => {
    try {
      await api.RevealFile(p);
    } catch {
      notify('That file is no longer there.');
    }
  };

  const sched = data?.schedule;
  const autoOn = (sched?.times?.length ?? 0) > 0 || !!sched?.instant;

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-mark">
            <img src={pouch} alt="NEST" />
          </div>
          <div>
            <div className="brand-name">NEST</div>
            <div className="brand-sub">File Organizer</div>
          </div>
        </div>

        {NAV.map((n) => (
          <div
            key={n.id}
            className={`nav-item ${page === n.id ? 'active' : ''}`}
            onClick={() => setPage(n.id)}
          >
            <Icon name={n.icon} size={18} />
            <span>{n.label}</span>
            {n.id === 'history' && history.length > 0 && (
              <span className="nav-badge">{history.length}</span>
            )}
          </div>
        ))}

        <div className="sidebar-foot">
          <div>
            <span className={`status-dot ${autoOn ? 'on' : 'off'}`} />
            {autoOn ? 'Automation on' : 'Manual only'}
          </div>
          <div style={{ marginTop: 6 }}>No background service.</div>
          <div className="credit" style={{ marginTop: 8 }}>
            Made by <strong>{about?.author ?? 'DERYK RYLOY'}</strong>
          </div>
        </div>
      </aside>

      <main className="main">
        <div className="topbar">
          <div>
            <h1 className="page-title">{TITLES[page].title}</h1>
            <div className="page-sub">{TITLES[page].sub}</div>
          </div>
          <div className="topbar-actions">
            <button
              className="btn theme-toggle"
              onClick={toggleTheme}
              title={cfg?.theme === 'light' ? 'Switch to dark' : 'Switch to light'}
            >
              <Icon name={cfg?.theme === 'light' ? 'moon' : 'sun'} size={17} />
            </button>
            <button className="btn" onClick={onPreview} disabled={busy}>
              <Icon name="eye" size={16} /> Preview
            </button>
            <button className="btn primary" onClick={onRun} disabled={busy}>
              {busy ? <span className="spinner" /> : <Icon name="sparkles" size={16} />}
              Organize Now
            </button>
          </div>
        </div>

        <div className="content">
          {page === 'dashboard' && (
            <Dashboard
              data={data}
              history={history}
              onOpen={(p) => {
                api.OpenFolder(p).catch(() => notify('That folder is no longer there.'));
              }}
            />
          )}
          {page === 'categories' && cfg && (
            <Categories
              categories={data?.categories ?? []}
              cfg={cfg}
              onSaved={refresh}
              notify={notify}
            />
          )}
          {page === 'scheduler' && cfg && data && about && (
            <Scheduler
              cfg={cfg}
              about={about}
              schedule={data.schedule}
              onSaved={refresh}
              notify={notify}
            />
          )}
          {page === 'history' && (
            <History
              history={history}
              onUndo={onUndo}
              onClear={onClearHistory}
              onReveal={onReveal}
            />
          )}
        </div>
      </main>

      {welcome && (
        <div className="overlay">
          <div className="card modal">
            <div className="modal-body" style={{ textAlign: 'center', paddingTop: 18 }}>
              <img className="welcome-mark" src={pouchBig} alt="NEST" />
              <h2 style={{ margin: '14px 0 2px' }}>Welcome to NEST</h2>
              <div className="page-sub">The file organizer that cleans up, then disappears.</div>
              <div className="welcome-points" style={{ textAlign: 'left' }}>
                <div className="welcome-point">
                  <Icon name="folder" size={17} />
                  <span>Pictures, videos, music and documents fly to their real Windows folders.</span>
                </div>
                <div className="welcome-point">
                  <Icon name="layers" size={17} />
                  <span>Leftovers get grouped — NEST creates a matching folder for each family.</span>
                </div>
                <div className="welcome-point">
                  <Icon name="bell" size={17} />
                  <span>Schedule daily times or sort new downloads instantly. One toast, then silence.</span>
                </div>
              </div>
              <div className="page-sub" style={{ margin: '14px 0' }}>
                Made by <strong style={{ color: 'var(--orange)' }}>{about?.author ?? 'DERYK RYLOY'}</strong>
              </div>
              <div className="row" style={{ justifyContent: 'center' }}>
                <button
                  className="btn primary"
                  onClick={() => {
                    dismissWelcome();
                    onRun();
                  }}
                >
                  <Icon name="sparkles" size={16} /> Organize Now
                </button>
                <button className="btn ghost" onClick={dismissWelcome}>
                  Explore first
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {report && (
        <PreviewModal reports={report} onClose={() => setReport(null)} onReveal={onReveal} />
      )}
      {toast && (
        <div className="toast">
          <Icon name="check" size={16} />
          {toast}
        </div>
      )}
    </div>
  );
}

function PreviewModal({
  reports,
  onClose,
  onReveal,
}: {
  reports: Report[];
  onClose: () => void;
  onReveal: (p: string) => void;
}) {
  const total = reports.reduce((n, r) => n + r.moves.filter((m) => !m.err).length, 0);
  const bytes = reports.reduce((n, r) => n + r.bytes, 0);
  return (
    <div className="overlay" onClick={onClose}>
      <div className="card modal" onClick={(e) => e.stopPropagation()}>
        <div className="card-pad row between">
          <div>
            <div className="section-title" style={{ margin: 0 }}>
              <Icon name="eye" size={18} /> Preview
            </div>
            <div className="page-sub">
              {total} file(s) · {formatBytes(bytes)}
            </div>
          </div>
          <button className="btn ghost small" onClick={onClose}>
            <Icon name="x" size={16} />
          </button>
        </div>
        <div className="modal-body">
          {reports.map((rep) => (
            <div key={rep.source} style={{ marginBottom: 18 }}>
              <div className="page-sub" style={{ marginBottom: 10 }}>
                {rep.source}
              </div>
              {rep.moves.filter((m) => !m.err).length === 0 && (
                <div className="empty">Nothing to move here.</div>
              )}
              {rep.moves
                .filter((m) => !m.err)
                .map((m, i) => (
                  <div
                    className="move-line"
                    key={i}
                    style={{ padding: '3px 0' }}
                    title={`${formatBytes(m.size)} · modified ${new Date((m.modified || 0) * 1000).toLocaleString()}`}
                  >
                    <Icon name={m.icon} size={14} />
                    <span>{baseName(m.from)}</span>
                    <span className="arrow">→</span>
                    <span style={{ flex: 1 }}>{shortPath(m.to)}</span>
                    <button
                      className="icon-btn"
                      title="Reveal the original file in Explorer"
                      onClick={() => onReveal(m.from)}
                    >
                      <Icon name="folder" size={14} />
                    </button>
                  </div>
                ))}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
