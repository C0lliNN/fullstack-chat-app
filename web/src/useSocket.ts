import { useEffect, useState } from 'react';
import { baseUrl } from './constants';

interface Message {
  data: string;
}

interface Input {
  chatCode: string;
  userName: string;
  onOpen?: () => void;
  onError?: () => void;
  onMessage?: (e: Message) => void;
  onClose?: () => void;
}

export default function useSocket(i: Input) {
  const [websocket, setWebsocket] = useState<WebSocket>();

  useEffect(() => {
    setWebsocket((ws) => {
      if (!ws) {
        ws = new WebSocket(
          `${baseUrl.replace('http', 'ws')}/chats?code=${i.chatCode}&user=${
            i.userName
          }`
        );
      }

      if (i.onOpen) {
        ws.addEventListener('open', i.onOpen);
      }
      if (i.onError) {
        ws.addEventListener('error', i.onError);
      }
      if (i.onMessage) {
        ws.addEventListener('message', i.onMessage);
      }
      if (i.onClose) {
        ws.addEventListener('close', i.onClose);
      }

      return ws;
    });
  }, [i.chatCode]);

  return websocket;
}
