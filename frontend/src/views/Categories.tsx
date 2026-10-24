import { useState } from 'react';
import Icon from '../components/Icon';
import * as api from '../../wailsjs/go/main/App';
import { config as models } from '../../wailsjs/go/models';
import type { AppConfig, Category, CustomRule } from '../types';

const KNOWN_TARGETS = ['pictures', 'videos', 'music', 'documents', 'downloads', 'desktop'];

// Mirrors the backend sanitiser: a target is a known-folder key or one folder
// name, never a path.
function cleanTarget(raw: string): string {
  const t = raw.trim().replace(/[\\/:*?"<>|]/g, '');
  if (!t || t === '.' || t === '..' || t === '_') return '';
  if (KNOWN_TARGETS.includes(t.toLowerCase())) return t.toLowerCase();
  return t;
}

export default function Categories({
  categories,
  cfg,
  onSaved,
  notify,
}: {
  categories: Category[];
  cfg: AppConfig;
  onSaved: () => void | Promise<void>;
  notify: (m: string) => void;
}) {
  const [rules, setRules] = useState<CustomRule[]>([...(cfg.customRules ?? [])]);
  const [ext, setExt] = useState('');
  const [target, setTarget] = useState('');
  const [saving, setSaving] = useState(false);

  const dirty =
    rules.length !== (cfg.customRules ?? []).length ||
    rules.some((r, i) => r.ext !== cfg.customRules?.[i]?.ext || r.target !== cfg.customRules?.[i]?.target);

  const addRule = () => {
    let e = ext.trim().toLowerCase();
    if (e && !e.startsWith('.')) e = '.' + e;
    if (e.length < 2 || e.includes('/') || e.includes('\\')) {
      notify('Type an extension first, like .blend');
      return;
    }
    const t = cleanTarget(target);
    if (!t) {
      notify('Give the rule a destination, like documents or _Blender');
      return;
    }
    if (rules.some((r) => r.ext === e)) {
      notify(`${e} already has a rule`);
      return;
    }
    setRules([...rules, { ext: e, target: t }]);
    setExt('');
    setTarget('');
  };

  const save = async () => {
    setSaving(true);
    try {
      await api.SaveConfig(models.Config.createFrom({ ...cfg, customRules: rules }));
      await onSaved();
      notify(rules.length ? `${rules.length} custom rule(s) saved` : 'Custom rules cleared');
    } catch (e) {
      notify('Could not save rules: ' + String(e));
    } finally {
      setSaving(false);
    }
  };

  return (
    <div>
      <div className="section-title">
        <Icon name="layers" size={17} /> Extension families
        <span className="count">{categories.length} groups</span>
      </div>
      <div className="grid grid-2">
        {categories.map((c) => (
          <div className="cat-row" key={c.key}>
            <div className="icon-badge">
              <Icon name={c.icon} size={18} />
            </div>
            <div style={{ flex: 1, minWidth: 0 }}>
              <div className="row between">
                <div className="name" style={{ fontWeight: 650 }}>
                  {c.name}
                </div>
                {c.known ? (
                  <span className="dest-tag known">Windows folder</span>
                ) : (
                  <span className="dest-tag cluster">Creates folder</span>
                )}
              </div>
              <div className="cat-exts">
                {c.exts.slice(0, 12).map((e) => (
                  <span className="ext-pill" key={e}>
                    {e}
                  </span>
                ))}
                {c.exts.length > 12 && <span className="ext-pill">+{c.exts.length - 12}</span>}
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="section-title">
        <Icon name="sparkles" size={17} /> Your rules
        <span className="count">always win</span>
      </div>
      <div className="card card-pad">
        <div className="stat-hint" style={{ lineHeight: 1.7, marginBottom: 12 }}>
          Point any extension at a Windows folder ({KNOWN_TARGETS.join(', ')}) or a new
          cluster folder like <span className="ext-pill">_Blender</span>. Rules beat the
          built-in families.
        </div>
        {rules.length === 0 && <div className="empty" style={{ padding: '14px' }}>No custom rules yet.</div>}
        {rules.map((r) => (
          <div className="slot-row" key={r.ext}>
            <span className="ext-pill">{r.ext}</span>
            <span className="arrow" style={{ color: 'var(--teal-soft)' }}>→</span>
            <span className="slot-time" style={{ fontSize: 13 }}>{r.target}</span>
            <span style={{ flex: 1 }} />
            <button
              className="btn ghost small"
              onClick={() => setRules(rules.filter((x) => x.ext !== r.ext))}
            >
              <Icon name="x" size={15} />
            </button>
          </div>
        ))}
        <div className="row wrap" style={{ marginTop: 12 }}>
          <input
            className="input"
            placeholder=".blend"
            value={ext}
            onChange={(e) => setExt(e.target.value)}
            style={{ maxWidth: 120 }}
          />
          <input
            className="input"
            placeholder="_Blender or documents"
            value={target}
            onChange={(e) => setTarget(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && addRule()}
            style={{ flex: 1, minWidth: 180 }}
            list="known-targets"
          />
          <datalist id="known-targets">
            {KNOWN_TARGETS.map((k) => (
              <option key={k} value={k} />
            ))}
          </datalist>
          <button className="btn small" onClick={addRule}>
            <Icon name="plus" size={15} /> Add rule
          </button>
        </div>
        <div className="row" style={{ marginTop: 14 }}>
          <button className="btn primary" onClick={save} disabled={saving || !dirty}>
            {saving ? <span className="spinner" /> : <Icon name="check" size={16} />} Save rules
          </button>
        </div>
      </div>

      <div className="section-title">
        <Icon name="sparkles" size={17} /> The leftover rule
      </div>
      <div className="card card-pad">
        <div className="stat-hint" style={{ lineHeight: 1.8 }}>
          Anything left over — extensions with no home — is grouped together. When 3 or more
          share a type, NEST creates a matching folder inside the source and moves them in.
          Stragglers drop into <span className="ext-pill">_Misc</span> so nothing is left loose.
        </div>
      </div>
    </div>
  );
}
