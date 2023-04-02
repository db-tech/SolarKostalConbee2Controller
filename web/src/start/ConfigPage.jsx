import {useEffect, useState} from 'react';
import {Button, Col, Container, Form, ListGroup, Modal, Row} from 'react-bootstrap';
import InitStatus from "../utils/InitStatus.js";

function ConfigPage({client, onStatusResponse}) {
    const [threshold, setThreshold] = useState(0);
    const [showModal, setShowModal] = useState(false);
    const [socketName, setSocketName] = useState('Please Wait');
    const [socketList, setSocketList] = useState([]);
    const [duration, setDuration] = useState(0);

    useEffect(() => {
        if (client === null) {
            return;
        }
        client.getProperties().then((properties) => {
            setThreshold(properties.Threshold);
            setSocketName(properties.PlugName);
            setDuration(properties.PollDuration)
        })
    }, [client])

    const handleThresholdChange = (e) => {
        setThreshold(e.target.valueAsNumber);
    };

    const handleDurationChange = (e) => {
        setDuration(e.target.valueAsNumber);
    }

    const handleSocketChange = (event) => {
        event.preventDefault();
        client.getLights().then((lights) => {
            console.log(lights);
            setSocketList(Object.values(lights));
        });
        setShowModal(true);
    };

    const handleSave = () => {
        client.saveProperties({Threshold: threshold, PlugName: socketName, PollDuration: duration}).then((response) => {
            onStatusResponse(response);
        });
    };

    function alertClicked(event) {
        event.preventDefault();
        console.log("alertClicked: " + event.target.textContent);
        setSocketName(event.target.textContent);
        setShowModal(false);
    }

    return (
        <Container fluid>
            <Form>
                <Form.Group as={Row} controlId="formThreshold">
                    <Form.Label column sm="4">Threshold</Form.Label>
                    <Col sm="8">
                        <Form.Control type="number" placeholder="Enter threshold in Watt" value={threshold}
                                      onChange={handleThresholdChange}/>
                    </Col>
                </Form.Group>

                <Form.Group as={Row} controlId="formDuration">
                    <Form.Label column sm="4">Duration (s)</Form.Label>
                    <Col sm="8">
                        <Form.Control type="number" placeholder="Enter duration in seconds" value={duration}
                                        onChange={handleDurationChange}/>
                    </Col>
                </Form.Group>

                <Form.Group as={Row} controlId="formSocket">
                    <Form.Label column sm="4">Socket</Form.Label>
                    <Col sm="6">
                        <Form.Control type="text" placeholder="Enter socket name" readOnly value={socketName}/>
                    </Col>
                    <Col sm="2">
                        <Button variant="primary" onClick={handleSocketChange}>Change</Button>
                    </Col>
                </Form.Group>
                <br/>
                <Form.Group as={Row} controlId="clientSettings">
                    <Col xs="6">
                        <Button variant="info" onClick={() => onStatusResponse({Status: InitStatus.DeconzAuth, StatusMessage: "Deconz Config"})}>Deconz Config</Button>
                    </Col>
                    <Col xs="6">
                        <Button variant="info" onClick={() => onStatusResponse({Status: InitStatus.KostalAuth, StatusMessage: "Kostal Config"})}>Kostal Config</Button>
                    </Col>
                </Form.Group>
                <br/>

                <div className="d-flex justify-content-center">
                    <Button variant="primary" onClick={handleSave}>Save</Button>
                </div>
            </Form>

            {showModal && (
                <Modal show={showModal} onHide={() => setShowModal(false)}>
                    <Modal.Header closeButton>
                        <Modal.Title>Change Socket</Modal.Title>
                    </Modal.Header>
                    <Modal.Body>
                        <ListGroup defaultActiveKey="#link1">
                            {socketList.map(socket => {
                                return <ListGroup.Item action onClick={alertClicked}>{socket.name}</ListGroup.Item>
                            })}

                        </ListGroup>
                    </Modal.Body>
                    <Modal.Footer>
                        <Button variant="secondary" onClick={() => setShowModal(false)}>
                            Close
                        </Button>
                        <Button variant="primary" onClick={() => setShowModal(false)}>
                            Save Changes
                        </Button>
                    </Modal.Footer>
                </Modal>
            )}
        </Container>
    );
}

export default ConfigPage;
