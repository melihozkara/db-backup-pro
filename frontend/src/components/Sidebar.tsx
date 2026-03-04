import { NavLink } from 'react-router-dom';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
import {
  LayoutDashboard,
  Database,
  History,
  Settings,
  Shield,
  ChevronLeft,
  ChevronRight,
  Github,
} from 'lucide-react';
import { cn } from '../lib/utils';

export default function Sidebar() {
  const [collapsed, setCollapsed] = useState(false);
  const { t } = useTranslation();

  const navigation = [
    { name: t('nav.dashboard'), path: '/', icon: LayoutDashboard },
    { name: t('nav.databases'), path: '/databases', icon: Database },
    { name: t('nav.history'), path: '/history', icon: History },
    { name: t('nav.settings'), path: '/settings', icon: Settings },
  ];

  return (
    <aside
      className={cn(
        'flex flex-col bg-zinc-900 border-r border-zinc-800 transition-all duration-300',
        collapsed ? 'w-16' : 'w-60'
      )}
    >
      {/* Logo */}
      <div className="h-14 flex items-center px-4 border-b border-zinc-800">
        <div className="flex items-center gap-3 overflow-hidden">
          <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center flex-shrink-0">
            <Shield className="w-4 h-4 text-white" />
          </div>
          {!collapsed && (
            <span className="text-sm font-bold text-white whitespace-nowrap">
              DB Backup Pro
            </span>
          )}
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-2 space-y-0.5 overflow-hidden">
        {navigation.map((item) => (
          <NavLink
            key={item.path}
            to={item.path}
            className={({ isActive }) =>
              cn(
                'flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors group',
                isActive
                  ? 'bg-blue-600/15 text-blue-400'
                  : 'text-zinc-400 hover:text-white hover:bg-zinc-800'
              )
            }
          >
            <item.icon className="w-4 h-4 flex-shrink-0" />
            {!collapsed && <span className="truncate">{item.name}</span>}
          </NavLink>
        ))}
      </nav>

      {/* Footer */}
      <div className="p-2 border-t border-zinc-800 space-y-0.5">
        <button
          onClick={() => {
            try { BrowserOpenURL('https://github.com/melihozkara'); } catch { window.open('https://github.com/melihozkara', '_blank'); }
          }}
          className="w-full flex items-center gap-3 px-3 py-2 rounded-lg text-zinc-500 hover:text-zinc-300 hover:bg-zinc-800 transition-colors cursor-pointer"
        >
          <Github className="w-4 h-4 flex-shrink-0" />
          {!collapsed && (
            <span className="text-xs text-left">
              <span className="block text-[10px] text-zinc-600">Developer</span>
              Melih Özkara
            </span>
          )}
        </button>
        <button
          onClick={() => setCollapsed(!collapsed)}
          className="w-full flex items-center justify-center gap-2 px-3 py-2 rounded-lg text-zinc-500 hover:text-zinc-300 hover:bg-zinc-800 transition-colors cursor-pointer"
        >
          {collapsed ? (
            <ChevronRight className="w-4 h-4" />
          ) : (
            <>
              <ChevronLeft className="w-4 h-4" />
              <span className="text-xs">{t('nav.collapse')}</span>
            </>
          )}
        </button>
      </div>
    </aside>
  );
}
