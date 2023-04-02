import {Button, Col, Container, Form, Image, Row} from 'react-bootstrap';
import {useState} from "react";
import logo from '../../public/logo.png';

function LoginPage({onLogin}) {
    const [username, setUsername] = useState('');

    const handleSubmit = e => {
        e.preventDefault();

        //check if username length is longer than 3
        if (username.length < 3) {
            alert('Username must be at least 3 characters long');
            return;
        } else {
            localStorage.setItem('user', username);
            window.location.reload();
        }
    };

    return (
        <div className="d-flex align-items-center" style={{height: '100vh'}}>
            <Container style={{
                maxWidth: '650px',
                border: '1px solid lightgray',
                backgroundColor: 'rgba(255, 255, 255, 0.1)',
            }}>
                <Row className="align-items-center">
                    <Col xs={2}>
                        <Image src={logo} width={80} fluid rounded/>
                    </Col>
                    <Col xs={10} className="text-center">
                        <h1>Kostal-ConbeeII Controller</h1>
                        <h4>Enter a username</h4>
                    </Col>
                </Row>
                <Row>
                    <Col xs={2}>
                    </Col>
                    <Col xs={10}>
                        <Form
                            onSubmit={handleSubmit}
                            className="mx-auto"
                            style={{maxWidth: '400px'}}>
                             <Form.Group controlId="formUsername">
                                <Form.Label>Username</Form.Label>
                                <Form.Control
                                    type="text"
                                    value={username}
                                    onChange={(e) => setUsername(e.target.value)}
                                />
                            </Form.Group>
                            <Button type="submit" onClick={onLogin}>Log In</Button>
                        </Form>
                    </Col>
                </Row>
            </Container>
        </div>
    );
}

export default LoginPage
