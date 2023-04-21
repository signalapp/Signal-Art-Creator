//
// Copyright 2023 Signal Messenger, LLC
// SPDX-License-Identifier: AGPL-3.0-only
//

import React, { lazy, Suspense, useEffect } from 'react';
import { useLoaderData, Outlet } from 'react-router-dom';
import { useDispatch } from 'react-redux';

import { type LoadLocaleResult } from '../util/i18n';
import { I18n } from '../contexts/I18n';
import { setCredentials } from '../slices/credentials';
import { useCredentials } from '../selectors/credentials';
import { isUnsupported } from '../util/browser';
import { parseCredentialsFromURL } from '../util/credentials';
import { UnsupportedBrowser } from './UnsupportedBrowser';

const LazySignIn = lazy(() => import('./SignIn'));

export function Root(): JSX.Element {
  const { locale, messages } = useLoaderData() as LoadLocaleResult;
  const credentials = useCredentials();
  const dispatch = useDispatch();

  useEffect(() => {
    const onHashChange = () => {
      dispatch(setCredentials(parseCredentialsFromURL()));
    };
    onHashChange();

    window.addEventListener('hashchange', onHashChange);
    return () => {
      window.removeEventListener('hashchange', onHashChange);
    };
  }, [dispatch]);

  if (messages.title && 'message' in messages.title) {
    document.title = messages.title.message ?? '';
  }

  let body: JSX.Element;

  if (isUnsupported()) {
    body = <UnsupportedBrowser />;
  } else if (!credentials) {
    body = (
      <Suspense>
        <LazySignIn />
      </Suspense>
    );
  } else {
    body = <Outlet />;
  }

  return (
    <I18n messages={messages} locale={locale}>
      {body}
    </I18n>
  );
}
