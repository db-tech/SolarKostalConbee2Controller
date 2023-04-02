import {Button, Card, Container, Form} from "react-bootstrap";
import {useEffect, useState} from "react";
import toast from "react-hot-toast";
import InitStatus from "../utils/InitStatus.js";

const DeconzAuthPage = ({client, onStatusResponse}) => {

    const [password, setPassword] = useState("");
    const [username, setUsername] = useState("");
    const [address, setAddress] = useState("");

    useEffect(() => {
        if (client === null) {
            return;
        }
        client.getProperties().then((properties) => {
            setAddress(properties.HostAddress);
        })
    }, [client])

    function handleSubmit(event) {
        event.preventDefault();
        client.authenticate(username, password, address).then((response) => {
            switch (response.Status) {
                case InitStatus.Ok:
                    console.log("InitStatus.Ok")
                    toast.success("Connected to Conbee2 Controller: " + response.StatusMessage);
                    break;
                case InitStatus.DeconzAuth:
                    console.log("InitStatus.Auth")
                    toast.error("Authentication failed: " + response.StatusMessage);
                    break;
                case InitStatus.Config:
                    console.log("InitStatus.Config")
                    toast.error("Configuration failed: " + response.StatusMessage);
                    break;
                case InitStatus.Error:
                    console.log("InitStatus.Error")
                    toast.error("Error: " + response.StatusMessage);
                    break;
                default:
                    console.log("default")
                    toast.error("Unknown error: " + response.StatusMessage);

            }
            onStatusResponse(response);
        });
    }

    return (
        <Container fluid>
            <div style={{height: "10px"}}></div>
            <Card>
                <Card.Header>Deconz Authentication</Card.Header>
                <Card.Body>

                    <Card.Text style={{fontSize: "14px"}}>
                        You can leave the username and password empty and authenticate with the Web-Ui.
                        Open the Gateway settings, select "Advanced" and click on the "Authenticate App" button.
                        Now press the "Authenticate" button at the bottom of this page.
                    </Card.Text>

                    <Form onSubmit={handleSubmit}>
                        <Form.Group className="mb-3" controlId="formBasicAddress">
                            <Form.Label>Address</Form.Label>
                            <Form.Control type="text"
                                            placeholder="Enter address"
                                            value={address}
                                            onChange={(e) => setAddress(e.target.value)}/>
                        </Form.Group>

                        <Form.Group className="mb-3" controlId="formBasicEmail">
                            <Form.Label>Username</Form.Label>
                            <Form.Control type="text"
                                          placeholder="Enter username"
                                          value={username}
                                          onChange={(e) => setUsername(e.target.value)}/>
                        </Form.Group>
                        <Form.Group className="mb-3" controlId="formBasicPassword">
                            <Form.Label>Password</Form.Label>
                            <Form.Control type="password" placeholder="Password"
                            value={password} onChange={(e) => setPassword(e.target.value)}/>
                        </Form.Group>
                        <Button variant="primary" type="submit">Authenticate</Button>
                    </Form>
                </Card.Body>
            </Card>
        </Container>
    )
}

export default DeconzAuthPage
