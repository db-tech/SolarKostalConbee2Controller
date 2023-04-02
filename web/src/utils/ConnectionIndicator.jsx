import {Badge} from "react-bootstrap";

// Define the ConnectionIndicator component
function ConnectionIndicator(props) {

    return (
        <>
            {props.isConnected && (
                <Badge bg="success" >
                    <span>{props.name} Connected</span>
                </Badge>
            )}
            {!props.isConnected && (
                <Badge  bg="danger">
                    <span>{props.name} Not Connected</span>
                </Badge>
            )}
        </>
    );
}

export default ConnectionIndicator;
