//
// Copyright 2023 Signal Messenger, LLC
// SPDX-License-Identifier: AGPL-3.0-only
//

import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { useDispatch } from 'react-redux';
import createDebug from 'debug';

import styles from './SignIn.module.scss';

import { Toaster } from '../components/Toaster';
import { PageHeader } from '../elements/PageHeader';
import { Spinner } from '../elements/Spinner';
import { useI18n } from '../contexts/I18n';
import { Provisioning } from '../util/provisioning';
import { ProvisioningToken } from '../util/protos';
import { noop } from '../util/noop';
import { MINUTE, SECOND } from '../constants';
import { addToast, dismissToast } from '../slices/art';
import { setCredentials } from '../slices/credentials';
import { useToasts } from '../selectors/art';

const debug = createDebug('signal:routes:SignIn');

const RECONNECT_DELAY = SECOND / 2;
const RECONNECT_DELAY_JITTER = SECOND / 4;
const RECONNECT_BASE = 1.2;
const MAX_RECONNECT_DELAY = 30 * SECOND;

function LinkToBlog(children: JSX.Element): JSX.Element {
  return (
    <a
      href="https://signal.org/blog/make-privacy-stick/"
      target="_blank"
      rel="noreferrer"
    >
      {children}
    </a>
  );
}

export default function SignIn(): JSX.Element {
  const i18n = useI18n();
  const dispatch = useDispatch();
  const toasts = useToasts();
  const [provisioning, setProvisioning] = useState<Provisioning | undefined>();
  const [token, setToken] = useState<string | undefined>();
  const [reconnectCounter, setReconnectCounter] = useState(0);
  const [delayedReconnectCounter, setDelayedReconnectCounter] = useState(0);
  const [retryCount, setRetryCount] = useState(0);
  const [isLinkOpen, setIsLinkOpen] = useState(false);

  const url = useMemo(() => {
    if (provisioning === undefined) {
      return undefined;
    }

    if (token === undefined) {
      return undefined;
    }

    return provisioning.getUrl(token);
  }, [token, provisioning]);

  // Rotate keys
  useEffect(() => {
    const update = async () => {
      setProvisioning(await Provisioning.create());
    };
    const timer = setInterval(update, 15 * MINUTE);
    update();

    return () => {
      clearInterval(timer);
    };
  }, []);

  // Delay reconnects
  useEffect(() => {
    let delay = Math.min(
      RECONNECT_DELAY * RECONNECT_BASE ** retryCount,
      MAX_RECONNECT_DELAY
    );
    delay += Math.round(Math.random() * RECONNECT_DELAY_JITTER);
    setToken(undefined);

    const timer = setTimeout(() => {
      setDelayedReconnectCounter(reconnectCounter);
    }, delay);

    debug('reconnecting after delay %dms', delay);

    return () => {
      clearTimeout(timer);
    };
  }, [retryCount, reconnectCounter]);

  const onEncryptedCredentials = useCallback(
    async (data: string | Uint8Array) => {
      try {
        if (!provisioning) {
          throw new Error('No local keys');
        }

        if (typeof data === 'string') {
          throw new Error('Unsupported provisioning response');
        }

        const message = await provisioning.decryptMessage(data);
        if (!message.username || !message.password) {
          throw new Error('Bad message');
        }
        dispatch(
          setCredentials({
            username: message.username,
            password: message.password,
          })
        );
      } catch (error) {
        if (!(error instanceof Error)) {
          return;
        }
        dispatch(
          addToast({
            key: 'StickerCreator--Toasts--errorSigningIn',
            subs: { message: error.message },
          })
        );
      }
    },
    [provisioning, dispatch]
  );

  useEffect(() => {
    if (!provisioning) {
      return noop;
    }

    const endpoint = new URL(document.location.href);
    endpoint.search = '';
    endpoint.hash = '';
    endpoint.protocol = endpoint.protocol === 'http:' ? 'ws:' : 'wss:';
    endpoint.pathname = '/api/socket';

    debug(
      'connecting websocket %s, counter=%d',
      endpoint,
      delayedReconnectCounter
    );

    const socket = new WebSocket(endpoint.toString());

    socket.binaryType = 'arraybuffer';

    let gotToken = false;

    socket.addEventListener(
      'message',
      ({ data }: MessageEvent<ArrayBuffer>) => {
        const bytes = new Uint8Array(data);

        if (!gotToken) {
          gotToken = true;

          const { token: maybeToken } = ProvisioningToken.decode(bytes);
          if (!maybeToken) {
            debug('no token');
            dispatch(
              addToast({
                key: 'StickerCreator--Toasts--errorSigningIn',
                subs: { message: 'No token' },
              })
            );
            setProvisioning(undefined);
            return;
          }

          setToken(maybeToken);
          return;
        }

        onEncryptedCredentials(bytes);
      }
    );

    let isClosed = false;

    const reconnect = () => {
      setReconnectCounter(counter => {
        // We are not in control
        if (counter !== delayedReconnectCounter) {
          return counter;
        }
        setRetryCount(count => count + 1);
        return counter + 1;
      });
    };

    socket.addEventListener('open', () => {
      debug('socket open');
      setRetryCount(0);
    });
    socket.addEventListener('error', () => {
      if (isClosed) {
        return;
      }
      debug('socket error');
      dispatch(
        addToast({
          key: 'StickerCreator--Toasts--errorSigningIn',
          subs: { message: 'Socket error' },
        })
      );
      reconnect();
    });
    socket.addEventListener('close', () => {
      if (isClosed) {
        return;
      }
      debug('socket closed');
      reconnect();
    });

    return () => {
      isClosed = true;
      socket.close();
    };
  }, [delayedReconnectCounter, onEncryptedCredentials, dispatch, provisioning]);

  const openInSignal = useCallback(
    (e: React.MouseEvent<HTMLButtonElement>) => {
      e.preventDefault();

      setIsLinkOpen(true);
    },
    [setIsLinkOpen]
  );

  useEffect(() => {
    if (!isLinkOpen) {
      return noop;
    }
    const timer = setTimeout(() => {
      setIsLinkOpen(false);
    }, 1000);
    return () => clearTimeout(timer);
  }, [isLinkOpen]);

  return (
    <>
      <div className={styles.container}>
        <PageHeader />
        <div className={styles.centered}>
          <h2 className={styles.title}>{i18n('SignIn--title')}</h2>

          <p className={styles.body}>
            {i18n.getIntl().formatMessage(
              { id: 'icu:SignIn--body' },
              {
                linkToBlog: LinkToBlog,
              }
            )}
          </p>

          {url ? (
            <button
              type="submit"
              className={styles.button}
              onClick={openInSignal}
            >
              {i18n('SignIn--link')}
            </button>
          ) : (
            <Spinner size={24} />
          )}
        </div>
      </div>
      {isLinkOpen && (
        <iframe
          title="Signal Link Opener"
          style={{ display: 'none' }}
          src={url}
        />
      )}
      <Toaster
        className={styles.toaster}
        loaf={toasts.map((slice, id) => ({
          id,
          text: i18n(slice.key, slice.subs),
        }))}
        onDismiss={() => dispatch(dismissToast())}
      />
    </>
  );
}
