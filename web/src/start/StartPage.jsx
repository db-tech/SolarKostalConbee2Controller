import React, {useEffect, useState} from 'react';
import {Button, Col, Container, Row, Table} from 'react-bootstrap';
import InitStatus from "../utils/InitStatus.js";
import toast from "react-hot-toast";

const StartPage = ({isConnected, client, onStatusResponse}) => {

    const [housePowerConsumption, setHousePowerConsumption] = useState(0)
    const [pvPowerGenerated, setPvPowerGenerated] = useState(0)
    const [gridOut, setGridOut] = useState(0)
    const [socketState, setSocketState] = useState(false)
    const [enabled, setEnabled] = useState(false)

    useEffect(() => {
        if (client == null) {
            console.log("localClient is null")
            return;
        }
        client.subscribe("data", (response) => {
            console.log("data: " + JSON.stringify(response))
            //round value with 2 digits
            const tmpHousePowerConsumption = Math.round(response.inverterData.HousePowerConsumption * 100) / 100
            const tmpPvPowerGenerated = Math.round(response.inverterData.PVPower * 100) / 100
            const tmpGridOut = Math.round(response.inverterData.Overproduction * 100) / 100

            setHousePowerConsumption(tmpHousePowerConsumption)
            setPvPowerGenerated(tmpPvPowerGenerated)
            setGridOut(tmpGridOut)
            setSocketState(response.socketState)
        })
        client.subscribe("monitoring", (response) => {
            console.log("monitoring: " + JSON.stringify(response))
            setEnabled(response.enabled)
        })
        client.status()
    }, [client])

    const switchSocketState = () => {
        client.getProperties().then((properties) => {
            if (socketState) {
                client.switchLightOff(properties.PlugName).then((response) => {
                    if (response.Status !== InitStatus.Ok) {
                        toast.error("Error: " + response.StatusMessage);
                    } else {
                        setSocketState(false);
                    }
                })
            } else {
                client.switchLightOn(properties.PlugName).then((response) => {
                    if (response.Status !== InitStatus.Ok) {
                        toast.error("Error: " + response.StatusMessage);
                    } else {
                        setSocketState(true);
                    }
                })
            }
        })
    }

    return (
        <div style={{backgroundColor: '#f8f9fa'}}>
            <Container fluid>
                <Row className="mt-3">
                    <Col xs={12} className="text-left">
                        <Table striped bordered hover>
                            <tbody>
                            <tr>
                                <td>PV Power Generated</td>
                                <td>{pvPowerGenerated} Watt</td>
                            </tr>
                            <tr>
                                <td>House Power Consumption</td>
                                <td>{housePowerConsumption} Watt</td>
                            </tr>
                            <tr>
                                <td>Overproduction</td>
                                <td style={{color: gridOut < 0 ? "red" : "green"}}>{gridOut} Watt</td>
                            </tr>
                            </tbody>
                        </Table>
                    </Col>
                </Row>

                <Row className="mt-3">
                    <Col md={12} className="text-center">
                        {socketState ? <Button variant="success" onClick={switchSocketState}>Socket: ON</Button> :
                            <Button variant="danger" onClick={switchSocketState}>Socket: OFF</Button>}
                    </Col>
                </Row>
            </Container>
            <div style={{height: "50px"}}/>


            <Container fluid>

                <Row className="mt-3">
                    <Col md={12} className="text-center">
                        <div className="btn-group-vertical btn-group-lg">
                            <Button onClick={() => onStatusResponse({Status: InitStatus.Config, StatusMessage: "Start Configuration"})}
                                    variant={"primary"} className="my-1">Configure</Button>
                            <Button variant={"primary"} className="my-1">Logs</Button>
                            {!enabled ? <Button onClick={() => client.startMonitoring()} variant={"secondary"}
                                                className="my-1">Enable</Button> :
                                <Button onClick={() => client.stopMonitoring()} variant={"primary"}
                                        className="my-1">Disable</Button>}
                        </div>

                    </Col>
                </Row>
            </Container>
        </div>
    );
};

export default StartPage;
