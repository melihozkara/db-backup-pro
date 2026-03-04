import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { GetSettings, SaveSettings, SetAppLanguage, TestTelegramConnection, SendTestTelegramMessage, GetServerConfig, SaveServerConfig } from '../api/client';
import { isWebMode } from '../api/mode';
import type { ServerConfig } from '../api/types';
import { useToast } from '../components/Toast';
import { PageLoader } from '../components/Spinner';
import PageHeader from '../components/PageHeader';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Select } from '../components/ui/select';
import { Switch } from '../components/ui/switch';
import { Separator } from '../components/ui/separator';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../components/ui/tabs';
import { Save, Loader2, MessageSquare, Wrench, Settings as SettingsIcon, Send, CheckCircle2, Globe } from 'lucide-react';

const defaultSettings: any = {
  telegram: { bot_token: '', chat_id: '', enabled: false },
  tool_paths: { pg_dump: '', mysqldump: '', mongodump: '' },
  default_retention: 7,
  language: 'tr',
};

export default function Settings() {
  const { t, i18n } = useTranslation();
  const [settings, setSettings] = useState<any>(defaultSettings);
  const [serverConfig, setServerConfig] = useState<ServerConfig>({ port: 8090, host: '127.0.0.1' });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testingToken, setTestingToken] = useState(false);
  const [sendingTest, setSendingTest] = useState(false);
  const toast = useToast();
  const webMode = isWebMode();

  useEffect(() => { loadSettings(); }, []);

  const loadSettings = async () => {
    try {
      const [data, srvCfg] = await Promise.all([
        GetSettings(),
        GetServerConfig().catch(() => ({ port: 8090, host: '127.0.0.1' })),
      ]);
      if (data) setSettings(data);
      if (srvCfg) setServerConfig(srvCfg);
    }
    catch { toast.error(t('common.loadingError'), t('settings.settingsLoadFailed')); }
    finally { setLoading(false); }
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      await SaveSettings(settings);
      await SaveServerConfig(serverConfig);
      // Dili aninda degistir (backend + frontend)
      await SetAppLanguage(settings.language);
      await i18n.changeLanguage(settings.language);
      toast.success(t('common.saved'), t('settings.settingsSaved'));
    }
    catch { toast.error(t('common.savingError'), t('settings.settingsSaveFailed')); }
    finally { setSaving(false); }
  };

  const handleTestToken = async () => {
    if (!settings.telegram.bot_token) { toast.warning(t('common.warning'), t('settings.telegram.enterBotToken')); return; }
    setTestingToken(true);
    try { await TestTelegramConnection(settings.telegram.bot_token); toast.success(t('settings.telegram.tokenValid'), t('settings.telegram.tokenVerified')); }
    catch (error) { toast.error(t('settings.telegram.tokenInvalid'), String(error)); }
    finally { setTestingToken(false); }
  };

  const handleSendTest = async () => {
    if (!settings.telegram.bot_token || !settings.telegram.chat_id) { toast.warning(t('common.warning'), t('settings.telegram.enterBotAndChat')); return; }
    setSendingTest(true);
    try { await SendTestTelegramMessage(settings.telegram.bot_token, settings.telegram.chat_id); toast.success(t('settings.telegram.messageSent'), t('settings.telegram.testMessageSent')); }
    catch (error) { toast.error(t('settings.telegram.sendFailed'), String(error)); }
    finally { setSendingTest(false); }
  };

  if (loading) return <PageLoader />;

  return (
    <div>
      <PageHeader title={t('settings.title')} description={t('settings.description')}>
        <Button onClick={handleSave} disabled={saving}>
          {saving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
          {saving ? t('common.saving') : t('common.save')}
        </Button>
      </PageHeader>

      <Tabs defaultValue="telegram">
        <TabsList>
          <TabsTrigger value="telegram">
            <MessageSquare className="w-4 h-4" /> {t('settings.tabTelegram')}
          </TabsTrigger>
          <TabsTrigger value="tools">
            <Wrench className="w-4 h-4" /> {t('settings.tabTools')}
          </TabsTrigger>
          <TabsTrigger value="general">
            <SettingsIcon className="w-4 h-4" /> {t('settings.tabGeneral')}
          </TabsTrigger>
        </TabsList>

        {/* Telegram Tab */}
        <TabsContent value="telegram">
          <Card className="border-zinc-800 bg-zinc-900">
            <CardHeader>
              <CardTitle>{t('settings.telegram.title')}</CardTitle>
              <CardDescription>{t('settings.telegram.description')}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-5">
              <div className="flex items-center gap-3">
                <Switch
                  checked={settings.telegram.enabled}
                  onCheckedChange={(checked) => setSettings({ ...settings, telegram: { ...settings.telegram, enabled: checked } })}
                />
                <Label className="cursor-pointer">{t('settings.telegram.enable')}</Label>
              </div>

              <Separator />

              <div className="space-y-2">
                <Label>{t('settings.telegram.botToken')}</Label>
                <div className="flex gap-2">
                  <Input
                    type="password"
                    className="flex-1"
                    placeholder="123456789:ABCdefGHIjklMNOpqrsTUVwxyz"
                    value={settings.telegram.bot_token}
                    onChange={(e) => setSettings({ ...settings, telegram: { ...settings.telegram, bot_token: e.target.value } })}
                  />
                  <Button variant="outline" onClick={handleTestToken} disabled={testingToken}>
                    {testingToken ? <Loader2 className="w-4 h-4 animate-spin" /> : <CheckCircle2 className="w-4 h-4" />}
                    {t('common.verify')}
                  </Button>
                </div>
                <p className="text-xs text-zinc-500">{t('settings.telegram.botTokenHelp')}</p>
              </div>

              <div className="space-y-2">
                <Label>{t('settings.telegram.chatId')}</Label>
                <div className="flex gap-2">
                  <Input
                    className="flex-1"
                    placeholder="-1001234567890"
                    value={settings.telegram.chat_id}
                    onChange={(e) => setSettings({ ...settings, telegram: { ...settings.telegram, chat_id: e.target.value } })}
                  />
                  <Button variant="outline" onClick={handleSendTest} disabled={sendingTest}>
                    {sendingTest ? <Loader2 className="w-4 h-4 animate-spin" /> : <Send className="w-4 h-4" />}
                    {t('common.test')}
                  </Button>
                </div>
                <p className="text-xs text-zinc-500">{t('settings.telegram.chatIdHelp')}</p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Tools Tab */}
        <TabsContent value="tools">
          <Card className="border-zinc-800 bg-zinc-900">
            <CardHeader>
              <CardTitle>{t('settings.tools.title')}</CardTitle>
              <CardDescription>{t('settings.tools.description')}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label>{t('settings.tools.pgDump')}</Label>
                <Input placeholder={t('settings.tools.pgDumpPlaceholder')} value={settings.tool_paths.pg_dump} onChange={(e) => setSettings({ ...settings, tool_paths: { ...settings.tool_paths, pg_dump: e.target.value } })} />
              </div>
              <div className="space-y-2">
                <Label>{t('settings.tools.mysqlDump')}</Label>
                <Input placeholder={t('settings.tools.mysqlDumpPlaceholder')} value={settings.tool_paths.mysqldump} onChange={(e) => setSettings({ ...settings, tool_paths: { ...settings.tool_paths, mysqldump: e.target.value } })} />
              </div>
              <div className="space-y-2">
                <Label>{t('settings.tools.mongoDump')}</Label>
                <Input placeholder={t('settings.tools.mongoDumpPlaceholder')} value={settings.tool_paths.mongodump} onChange={(e) => setSettings({ ...settings, tool_paths: { ...settings.tool_paths, mongodump: e.target.value } })} />
              </div>

              <Separator />

              <div className="bg-zinc-800/50 rounded-lg p-4 border border-zinc-800">
                <h4 className="text-xs font-medium text-zinc-300 mb-2">{t('settings.tools.installGuide')}</h4>
                <div className="text-xs text-zinc-500 space-y-1">
                  <p><span className="text-zinc-400">PostgreSQL:</span> brew install postgresql (macOS) | apt install postgresql-client (Linux)</p>
                  <p><span className="text-zinc-400">MySQL:</span> brew install mysql-client (macOS) | apt install mysql-client (Linux)</p>
                  <p><span className="text-zinc-400">MongoDB:</span> brew install mongodb-database-tools (macOS) | apt install mongodb-database-tools (Linux)</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* General Tab */}
        <TabsContent value="general">
          <Card className="border-zinc-800 bg-zinc-900">
            <CardHeader>
              <CardTitle>{t('settings.general.title')}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label>{t('settings.general.defaultRetention')}</Label>
                <Input type="number" min={1} max={365} value={settings.default_retention} onChange={(e) => setSettings({ ...settings, default_retention: parseInt(e.target.value) || 7 })} />
                <p className="text-xs text-zinc-500">{t('settings.general.defaultRetentionDesc')}</p>
              </div>

              <div className="space-y-2">
                <Label>{t('settings.general.language')}</Label>
                <Select value={settings.language} onChange={(e) => setSettings({ ...settings, language: e.target.value })}>
                  <option value="tr">{t('settings.general.turkish')}</option>
                  <option value="en">{t('settings.general.english')}</option>
                </Select>
              </div>

              <Separator />

              {/* Server Config (Web Mode) */}
              <div className="space-y-3">
                <div className="flex items-center gap-2">
                  <Globe className="w-4 h-4 text-blue-400" />
                  <h4 className="text-sm font-medium text-zinc-300">{t('settings.server.title')}</h4>
                </div>
                <p className="text-xs text-zinc-500">{t('settings.server.description')}</p>
                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-2">
                    <Label>{t('settings.server.host')}</Label>
                    <Input
                      placeholder="127.0.0.1"
                      value={serverConfig.host}
                      onChange={(e) => setServerConfig({ ...serverConfig, host: e.target.value })}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>{t('settings.server.port')}</Label>
                    <Input
                      type="number"
                      min={1}
                      max={65535}
                      value={serverConfig.port}
                      onChange={(e) => setServerConfig({ ...serverConfig, port: parseInt(e.target.value) || 8090 })}
                    />
                  </div>
                </div>
                {webMode && (
                  <p className="text-[10px] text-amber-400/80">{t('settings.server.restartRequired')}</p>
                )}
              </div>

              <Separator />

              <div>
                <h4 className="text-sm font-medium text-zinc-300 mb-3">{t('settings.general.aboutApp')}</h4>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-zinc-500">{t('settings.general.version')}</span>
                    <span className="text-zinc-300">1.0.0</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-zinc-500">{t('settings.general.developer')}</span>
                    <span className="text-zinc-300">DB Backup Pro Team</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-zinc-500">{t('settings.general.dataDirectory')}</span>
                    <span className="text-zinc-300 font-mono text-xs">~/.dbbackup/</span>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
