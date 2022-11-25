import { Button, Col, Divider, Input, Row, Typography } from 'antd';
import axios from 'axios';
import { useState } from 'react';
import { useNavigate } from 'react-router';
import './App.css';
import { baseUrl } from './constants';

const { Text, Title } = Typography;

function App() {
  const [code, setCode] = useState('');
  const [userName, setUsername] = useState('');
  const navigate = useNavigate();

  async function handleCreateNewRoom() {
    try {
      const response = await axios.post(`${baseUrl}/chats`);
      handleJoinChatRoom(response.data.Code);
    } catch (e) {
      alert(e);
    }
  }

  function handleJoinChatRoom(code: string): void {
    navigate(`/chats/${code}?user=${userName}`);
  }

  return (
    <div className="App">
      <Title>Chat App</Title>
      <Row>
        <Col sm={24}>
          <Input
            size="large"
            placeholder="Enter your username"
            value={userName}
            onChange={(e) => setUsername(e.target.value)}
          />
        </Col>
      </Row>
      <Divider />
      <Button type="primary" onClick={handleCreateNewRoom}>
        Create new Room
      </Button>
      <Text className="Or">Or</Text>
      <Row>
        <Col sm={24}>
          <Input
            size="large"
            placeholder="Enter your code"
            value={code}
            onChange={(e) => setCode(e.target.value)}
          />
        </Col>
      </Row>

      <Button
        type="primary"
        className="Join"
        onClick={() => handleJoinChatRoom(code)}
      >
        Join Existing Room
      </Button>
    </div>
  );
}

export default App;
