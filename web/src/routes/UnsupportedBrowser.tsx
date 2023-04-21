//
// Copyright 2023 Signal Messenger, LLC
// SPDX-License-Identifier: AGPL-3.0-only
//

import React from 'react';
import styles from './UnsupportedBrowser.module.scss';

import { useI18n } from '../contexts/I18n';

export function UnsupportedBrowser(): JSX.Element {
  const i18n = useI18n();
  return (
    <div className={styles.container}>
      {i18n('UnsupportedBrowser--description')}
    </div>
  );
}
