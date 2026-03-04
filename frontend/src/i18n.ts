import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import { GetTranslations, GetSettings } from './api/client';

// i18next baslatici - Wails backend'den ceviri yukler
export async function initI18n() {
  // Kaydedilmis dili ve cevirileri paralel yukle
  let savedLang = 'tr';
  let trTranslations = {};
  let enTranslations = {};

  try {
    const [settings, tr, en] = await Promise.all([
      GetSettings(),
      GetTranslations('tr'),
      GetTranslations('en'),
    ]);
    if (settings?.language) savedLang = settings.language;
    trTranslations = tr;
    enTranslations = en;
  } catch (error) {
    console.error('Failed to load translations:', error);
  }

  await i18n.use(initReactI18next).init({
    resources: {
      tr: { translation: trTranslations },
      en: { translation: enTranslations },
    },
    lng: savedLang,
    fallbackLng: 'tr',
    interpolation: {
      escapeValue: false, // React zaten XSS koruyor
    },
  });

  return i18n;
}

export default i18n;
