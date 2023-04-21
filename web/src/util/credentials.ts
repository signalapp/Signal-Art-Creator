//
// Copyright 2023 Signal Messenger, LLC
// SPDX-License-Identifier: AGPL-3.0-only
//

import type { Credentials } from '../types.d';

export function parseCredentialsFromURL(): Credentials | undefined {
  const match = document.location.hash.match(/auth=(.+):([^:]+)($|&)/);
  if (!match) {
    return undefined;
  }

  const [, username, password] = match;
  return { username, password };
}
