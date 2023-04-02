import {Col, Container, Row} from "react-bootstrap";
import ConnectionIndicator from "../utils/ConnectionIndicator.jsx";
import React from "react";


const Title = ({isConnected, title}) => {
    return (
        <div className="title">
            <Container fluid>
                <Row className="mt-0">
                    <Col xs={12} className="text-center pt-3" style={{backgroundColor: 'black', color: '#fff'}}>
                        <h1 style={{margin: 0}}>
                            <p>{title}</p>
                            <ConnectionIndicator isConnected={isConnected} name="Controller"/>
                        </h1>

                    </Col>
                </Row>
            </Container>
        </div>
    );
}

export default Title;