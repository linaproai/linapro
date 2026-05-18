import Image from '@tiptap/extension-image';
import Link from '@tiptap/extension-link';
import Placeholder from '@tiptap/extension-placeholder';
import Underline from '@tiptap/extension-underline';
import StarterKit from '@tiptap/starter-kit';

import { $t } from '#/locales';

export function getExtensions(placeholder?: string) {
  return [
    StarterKit,
    Underline,
    Image.configure({
      inline: true,
      allowBase64: true,
    }),
    Link.configure({
      openOnClick: false,
    }),
    Placeholder.configure({
      placeholder: placeholder || $t('pages.editor.placeholder'),
    }),
  ];
}
