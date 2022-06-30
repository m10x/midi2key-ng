import {useState} from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import {Initialize} from "../wailsjs/go/main/App";
import {LoadDevices} from "../wailsjs/go/main/App";
import {Listen} from "../wailsjs/go/main/App";


function App() { 

    const lblStatus = "Please choose your device below 👇";
    const lblDevice = "";

    const startListen = "Start Listening!";
    const stopListen = "Stop Listening!";
    const noDeviceFound = "No Device Found. Please (double) Refresh";

    const [listenText, setListenText] = useState(startListen);

    const [chosenDevice, setChosenDevice] = useState("");
    const updateChosenDevice = (e) => setChosenDevice(e.target.value);

    const [devices, setDevices] = useState("");
    const updateDevices = (result) => setDevices(result);

    const [options, setOptions] = useState([
        { label: noDeviceFound, value: noDeviceFound },
    ]);

    function initialize() {
        Initialize();
    }

    function loadDevices() {
        LoadDevices().then(updateDevices);
        var deviceArr = devices.split("]");
        let deviceBuffer = []
        for (var i=deviceArr.length-1; i > 0; i--) { // reverse order, so [0] MIDI TROUGH is last and not the default selected
            var device = deviceArr[i].split(":")[0]
            device = device.trim();
            if (device == "RtMidi Output Client") {
                console.log("Skipping RtMidi Output Client");
                continue;
            }
            deviceBuffer.push({ label: device, value: device });
        }
        if (deviceBuffer.length > 0) {
            setOptions(
                deviceBuffer.map((buffer) => (
                    { label: buffer.label, value: buffer.value }
                ))
            );
            document.getElementById("btnChangeListen").className = "btn";
            setChosenDevice(deviceBuffer[0].label); //set chosen device to the first one in the buffer
            document.getElementById("selDevice").selectedIndex = 0;
        } else {
            setOptions([{ label: noDeviceFound, value: noDeviceFound }]);
            document.getElementById("btnChangeListen").className = "btn-disabled";
        }
    }

    // https://stackoverflow.com/questions/38558200/react-setstate-not-updating-immediately
    function listen() {
        if (listenText == startListen) {
            Listen(true, chosenDevice)
            setListenText(stopListen);
            document.getElementById("selDevice").className = "select-disabled";
            document.getElementById("btnRefresh").className = "btn-disabled";
        } else {
            Listen(false, chosenDevice)
            setListenText(startListen);
            document.getElementById("selDevice").className = "select";
            document.getElementById("btnRefresh").className = "btn";
        }
    }

    return (
        <div id="App" onLoad={initialize}>
            <img src={logo} id="logo" alt="logo"/>
            <div className="result">{lblStatus}</div>
            <div id="dropdown" className="input-box">
                <label>{lblDevice}
                    <select id="selDevice" onChange={updateChosenDevice} className="select">
                        {
                            options.map((option) => (
                                <option value={option.value}>{option.label}</option>
                            ))
                        }
                    </select>
                </label>
            </div>
            <div id="start" className="input-box">
                <button id="btnRefresh" className="btn" onClick={loadDevices}>Refresh Devices</button>
                <button id="btnChangeListen" className="btn-disabled" onClick={listen}>{listenText}</button>
            </div>
        </div>
    )
}

export default App
