//
// Copyright 2023 Signal Messenger, LLC
// SPDX-License-Identifier: AGPL-3.0-only
//

import createDebug from 'debug';

import { loadImage } from './loadImage';
// eslint-disable-next-line import/no-unresolved
import WebpWorker from './webpWorker?worker';
import type { Request, Response, Options } from './webpWorker';

const debug = createDebug('signal:util:webp');

let worker: Worker | undefined;

const queue = new Array<(response: Response) => void>();

function getWorker(): Worker {
  if (worker) {
    return worker;
  }

  debug('creating worker');
  worker = new WebpWorker();

  worker.addEventListener('message', message => {
    const response = message.data as Response;

    queue.shift()?.(response);
  });

  return worker;
}

export async function encode(
  data: Uint8Array,
  options: Options
): Promise<Uint8Array> {
  const image = await loadImage(data);

  const canvas = document.createElement('canvas');
  canvas.width = image.naturalWidth;
  canvas.height = image.naturalHeight;

  const ctx = canvas.getContext('2d');
  if (!ctx) {
    throw new Error('Failed to get 2d context from canvas');
  }

  ctx.drawImage(image, 0, 0);
  const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height);

  const webpDataPromise = new Promise<Uint8Array>((resolve, reject) => {
    queue.push(response => {
      if (response.error !== undefined) {
        debug('got worker error', response.error);
        reject(response.error);
      } else {
        debug('got worker response');
        resolve(response.data);
      }
    });
  });

  const request: Request = {
    imageData,
    options,
  };

  debug('sending worker request');
  getWorker().postMessage(request);

  return webpDataPromise;
}
