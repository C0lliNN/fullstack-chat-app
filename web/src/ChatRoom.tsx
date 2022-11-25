import { Button, Card, Input, Typography } from 'antd';
import { useMemo, useState } from 'react';
import { useParams } from 'react-router';
import { useSearchParams } from 'react-router-dom';
import './ChatRoom.css';
import useSocket from './useSocket';

const { Title } = Typography;
const { TextArea } = Input;

export default function ChatRoom() {
  const { chatCode } = useParams();
  const [searchParams] = useSearchParams();
  const [messages, setMessages] = useState([]);
  const [content, setContent] = useState('');
  const user = useMemo(() => searchParams.get('user'), [searchParams]);

  const websocket = useSocket({
    onOpen: handleOpenSocket,
    onMessage: handleNewMessage,
    chatCode: chatCode as string,
    userName: user as string
  });

  function handleOpenSocket() {
    console.log('Open Socket');
  }

  function handleSendMessage() {
    websocket?.send(JSON.stringify({ content }));
    setContent('');
  }

  function handleNewMessage(e: any) {
    console.log('New Message');
    setMessages((msgs) => msgs.concat(JSON.parse(e.data) as never));
  }

  // Do this
  return (
    <div className="ChatRoomContainer">
      <Title>Chat Room: {chatCode}</Title>
      <Title level={4}>User name: {user}</Title>

      <div className="Messages">
        {messages.map((m: any) => (
          <div key={m.ID} className="MessageContainer">
            <Card
              title={m.User.Name}
              className="MessageCard"
              style={{
                marginLeft: m.User.Name === user ? '0' : 'auto'
              }}
            >
              <p>{m.Content}</p>
            </Card>
          </div>
        ))}
      </div>

      <div className="SendForm">
        <TextArea
          size="large"
          rows={3}
          placeholder="Enter your message"
          value={content}
          onChange={(e) => setContent(e.target.value)}
        />
        <Button
          type="primary"
          className="SendButton"
          onClick={() => handleSendMessage()}
        >
          Send
        </Button>
      </div>
    </div>
  );
}
