import {useState} from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import {Greet} from "../wailsjs/go/main/App";
import {LoadDevices} from "../wailsjs/go/main/App";

function App() { 
    const [resultText, setResultText] = useState("Please enter your name below 👇");
    const [name, setName] = useState('');
    const updateName = (e) => setName(e.target.value);
    const updateResultText = (result) => setResultText(result);
    const updateDevices = (result) => setDevices(result);
    const [devices, setDevices] = useState("");

    const [options, setOptions] = useState([
        { label: "No Device Found. Please (double) Refresh", value: "none" },
    ]);
    const updateOptions = (result) => setOptions(result);


    function greet() {
        Greet(name).then(updateResultText);
    }

    function loadDevices() {
        LoadDevices().then(updateDevices);
        //console.log("devices", devices);
        var deviceArr = devices.split("]");
        //console.log("deviceArr", deviceArr);
        let deviceBuffer = []
        for (var i=1; i < deviceArr.length; i++) {
            //console.log("device", deviceArr[i]);
            var device = deviceArr[i].split(":")[0]
            device = device.trim();
            deviceBuffer.push({ label: device, value: device });
        }
        console.log(deviceBuffer);
        if (deviceBuffer.length > 0) {
            setOptions(
                deviceBuffer.map((buffer) => (
                { label: buffer.label, value: buffer.value }
                ))
            );
        } else {
            setOptions([{ label: "No Device Found. Please Refresh", value: "none" }]);
        }
    }

    /*
                <div id="result" className="result">{resultText}</div>
            <div id="input" className="input-box">
                <input id="name" className="input" onChange={updateName} autoComplete="off" name="input" type="text"/>
                <button className="btn" onClick={greet}>Greet</button>
            </div>
    */

    return (
        <div id="App">
            <img src={logo} id="logo" alt="logo"/>
            <div className="result">Please choose your device below 👇</div>
            <div id="dropdown" className="input-box">
                <label>Which Device shall be used?
                    <select onLoad={loadDevices}>
                        {
                            options.map((option) => (
                                <option value={option.value}>{option.label}</option>
                            ))
                        }
                    </select>
                </label>
            </div>
            <div id="start" className="input-box">
                <button className="btn" onClick={loadDevices}>Refresh Devices</button>
                <button className="btn" onClick={loadDevices}>Go!</button>
            </div>
        </div>
    )
}

export default App
