//
// Copyright 2023 Signal Messenger, LLC
// SPDX-License-Identifier: AGPL-3.0-only
//

import { encode } from '@jsquash/webp';

export type Options = Readonly<{
  quality: number;
}>;

export type Request = Readonly<{
  imageData: ImageData;
  options: Options;
}>;

export type Response = Readonly<
  | {
      error: string;
      data?: undefined;
    }
  | {
      error?: undefined;
      data: Uint8Array;
    }
>;

// eslint-disable-next-line no-restricted-globals
addEventListener('message', async message => {
  const request: Request = message.data;

  let response: Response;
  try {
    const arrayBuffer = await encode(request.imageData, request.options);
    response = { data: new Uint8Array(arrayBuffer) };
  } catch (error) {
    response = { error: String(error) };
  }
  postMessage(response);
});
