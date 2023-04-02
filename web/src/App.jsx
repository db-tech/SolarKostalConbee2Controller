import {useEffect, useState} from 'react'
import './App.css'
import {JRpcWsClient} from "./JRpcWsClient.js";
import StartPage from "./start/StartPage.jsx";
import Title from "./start/Title.jsx";
import DeconzAuthPage from "./start/DeconzAuthPage.jsx";
import toast, {Toaster} from 'react-hot-toast';
import InitStatus from "./utils/InitStatus.js";
import ConfigPage from "./start/ConfigPage.jsx";
import KostalAuthPage from "./start/KostalAuthPage.jsx";
import ErrorPage from "./start/ErrorPage.jsx";
import LoginPage from "./start/LoginPage.jsx";


function App() {
    const [count, setCount] = useState(0)
    const [localClient, setLocalClient] = useState(null)
    const [initStatus, setInitStatus] = useState(InitStatus.Ok);
    const [statusResponseMsg, setStatusResponseMsg] = useState(null);
    const [title, setTitle] = useState("Start Page");

    useEffect(() => {
        const host = window.location.hostname;
        const client = new JRpcWsClient({
            url: `ws://${host}:8888/ws`,
            name: 'Conbee2Controller',
            SolarKostalConbee2Controller: (i) => {
                if (i) {
                    console.log("JRpcWsClient Connected")
                    // client.login("admin" + props.user);
                    setLocalClient(client)
                } else {
                    console.log("JRpcWsClient Disconnected")
                    setLocalClient(null)
                }
            }
        });
    }, []);


    const showPrimaryToast = (message, duration) => {

        toast(message, {
            duration: duration || 4000,
            //Exclamation mark
            icon: '❕',
            style: {
                borderRadius: '4px',
                background: '#007bff',
                color: '#fff',
            },
            iconTheme: {
                primary: '#fff',
                secondary: '#007bff',
            },
        });
    };

    function statusResponse(response) {
        console.log("Status: " + JSON.stringify(response))
        console.log(response)
        setInitStatus(response.Status)
        setStatusResponseMsg(response)
        switch (response.Status) {
            case InitStatus.Ok:
                setTitle("Start Page");
                toast.success("Connected to Conbee2 Controller");
                break;
            case InitStatus.DeconzAuth:
                setTitle("Deconz Authentication");
                showPrimaryToast(response.StatusMessage);
                break;
            case InitStatus.Config:
                setTitle("Configuration");
                showPrimaryToast(response.StatusMessage);
                break;
            case InitStatus.KostalAuth:
                setTitle("Kostal Authentication");
                showPrimaryToast(response.StatusMessage);
                break;
            case InitStatus.Error:
                setTitle("Error");
                toast.error("Error: " + response.StatusMessage);
                break;
            default:
                setTitle("Start Page");
        }
    }

    function login(username) {
        if (localClient == null) {
            console.log("localClient is null")
            return;
        }
        localClient.login(username)
        localClient.status().then((response) => {
            statusResponse(response)
        })
    }

    useEffect(() => {
        if (localClient == null) {
            console.log("localClient is null")
            return;
        }
        if (localStorage.getItem("user") !== null) {
            login(localStorage.getItem("user"))
        }
    }, [localClient])


    return (
        <div>
            <Title isConnected={localClient !== null} title={title}/>
            <Toaster
                reverseOrder={true}
            />
            {(initStatus ===InitStatus.Login || localStorage.getItem("user") === null) && <LoginPage onLogin={login}/>}
            {initStatus === InitStatus.Ok && <StartPage client={localClient} isConnected={localClient !== null} onStatusResponse={statusResponse}/>}
            {initStatus === InitStatus.DeconzAuth &&
                <DeconzAuthPage client={localClient} onStatusResponse={statusResponse}/>}
            {initStatus === InitStatus.Config && <ConfigPage client={localClient} onStatusResponse={statusResponse}/>}
            {initStatus === InitStatus.KostalAuth && <KostalAuthPage client={localClient} onStatusResponse={statusResponse}/>}
            {initStatus === InitStatus.Error && <ErrorPage response={statusResponseMsg}/>}
        </div>
    )
}

export default App
