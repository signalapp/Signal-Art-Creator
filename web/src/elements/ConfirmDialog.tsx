//
// Copyright 2023 Signal Messenger, LLC
// SPDX-License-Identifier: AGPL-3.0-only
//

import React from 'react';

import { useI18n } from '../contexts/I18n';
import styles from './ConfirmDialog.module.scss';
import { Button } from './Button';

export type Props = Readonly<{
  title: string;
  children: React.ReactNode;
  confirm: string;
  onConfirm: () => unknown;
  cancel?: string;
  onCancel: () => unknown;
}>;

export function ConfirmDialog({
  title,
  children,
  confirm,
  cancel,
  onConfirm,
  onCancel,
}: Props): JSX.Element {
  const i18n = useI18n();
  const cancelText = cancel || i18n('StickerCreator--ConfirmDialog--cancel');

  return (
    <div className={styles.base}>
      <h1 className={styles.title}>{title}</h1>
      <p className={styles.text}>{children}</p>
      <div className={styles.grow} />
      <div className={styles.bottom}>
        <Button onClick={onCancel}>{cancelText}</Button>
        <Button primary onClick={onConfirm}>
          {confirm}
        </Button>
      </div>
    </div>
  );
}
