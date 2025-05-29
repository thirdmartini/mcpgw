
/*
var audioElementSource = document.getElementsByClassName("audio-element")[0]
    .getElementsByTagName("source")[0];
*/

export function PlayOnce(id, rawWav) {
    const blob = new Blob([rawWav], {type: "audio/wav"});
    const blobUrl = URL.createObjectURL(blob);

    //console.log("Playing audio:", blobUrl);
    //Play("player", blob);

    const audio = new Audio();
    audio.src = blobUrl;
    audio.controls = true;
    document.body.appendChild(audio);
    audio.play();
}


export function Play(id, recorderAudioAsBlob) {
    //read content of files (Blobs) asynchronously
    let reader = new FileReader();

    //once content has been read
    reader.onload = (e) => {
        //store the base64 URL that represents the URL of the recording audio
        let base64URL = e.target.result;

        //If this is the first audio playing, create a source element
        //as pre populating the HTML with a source of empty src causes error
        //if (!audioElementSource) //if its not defined create it (happens first time only)
        //    createSourceForAudioElement();


        let audioElement= document.getElementById(id);
        while (audioElement.firstChild) {
            audioElement.removeChild(audioElement.firstChild);
        }

        let audioElementSource = document.createElement("source");
        audioElement.appendChild(audioElementSource);



        //set the audio element's source using the base64 URL
        audioElementSource.src = base64URL;

        //set the type of the audio element based on the recorded audio's Blob type
        let BlobType = recorderAudioAsBlob.type.includes(";") ?
            recorderAudioAsBlob.type.substr(0, recorderAudioAsBlob.type.indexOf(';')) : recorderAudioAsBlob.type;
        audioElementSource.type = BlobType

        //call the load method as it is used to update the audio element after changing the source or other settings
        audioElement.load();

        //play the audio after successfully setting new src and type that corresponds to the recorded audio
        console.log("Playing audio...");
        audioElement.play();

        //Display text indicator of having the audio play in the background
        //displayTextIndicatorOfAudioPlaying();
    };

    //read content and convert it to a URL (base64)
    reader.readAsDataURL(recorderAudioAsBlob);

}


export function NewRecorder() {
    let audioRecorder = {
        /** Stores the reference of the MediaRecorder instance that handles the MediaStream when recording starts*/
        mediaRecorder: null, /*of type MediaRecorder*/
        /** Stores the reference to the stream currently capturing the audio*/
        streamBeingCaptured: null, /*of type MediaStream*/
        audioBlobs: [],

        /** Start recording the audio
         * @returns {Promise} - returns a promise that resolves if audio recording successfully started
         */
        Start: function ( onSegmentCaptured, options ) {
            //Feature Detection


            if (!(navigator.mediaDevices && navigator.mediaDevices.getUserMedia)) {
                //Feature is not supported in browser
                //return a custom error
                return Promise.reject(new Error('mediaDevices API or getUserMedia method is not supported in this browser.'));
            } else {
                //Feature is supported in browser
                //create an audio stream
                return navigator.mediaDevices.getUserMedia({audio: true}/*of type MediaStreamConstraints*/)
                    //returns a promise that resolves to the audio stream
                    .then(stream /*of type MediaStream*/ => {

                        audioRecorder.audioBlobs = []
                        //save the reference of the stream to be able to stop it when necessary
                        audioRecorder.streamBeingCaptured = stream;

                        //create a media recorder instance by passing that stream into the MediaRecorder constructor
                        audioRecorder.mediaRecorder = new MediaRecorder(stream, options);

                        //add a dataavailable event listener in order to store the audio data Blobs when recording
                        audioRecorder.mediaRecorder.addEventListener("dataavailable", event => {
                            //store audio Blob object
                            onSegmentCaptured(event.data)
                            audioRecorder.audioBlobs.push(event.data);
                        });

                        //start the recording by calling the start method on the media recorder
                        audioRecorder.mediaRecorder.start();
                    });
            }
        },

        /** Stop the started audio recording
         * @returns {Promise} - returns a promise that resolves to the audio as a blob file
         */
        Stop: function () {
            //return a promise that would return the blob or URL of the recording
            return new Promise(resolve => {
                //save audio type to pass to set the Blob type
                let mimeType = audioRecorder.mediaRecorder.mimeType;

                //listen to the stop event in order to create & return a single Blob object
                audioRecorder.mediaRecorder.addEventListener("stop", () => {
                    //create a single blob object, as we might have gathered a few Blob objects that needs to be joined as one
                    let audioBlob = new Blob(audioRecorder.audioBlobs, {type: mimeType});

                    //resolve promise with the single audio blob representing the recorded audio
                    resolve(audioBlob);
                });
                audioRecorder.Cancel();
            });
        },

        /** Cancel audio recording*/
        Cancel: function () {
            //stop the recording feature
            audioRecorder.mediaRecorder.stop();

            //stop all the tracks on the active stream in order to stop the stream
            audioRecorder.stopStream();

            //reset API properties for next recording
            audioRecorder.resetRecordingProperties();
        },

        /** Stop all the tracks on the active stream in order to stop the stream and remove
         * the red flashing dot showing in the tab
         */
        stopStream: function () {
            //stopping the capturing request by stopping all the tracks on the active stream
            audioRecorder.streamBeingCaptured.getTracks() //get all tracks from the stream
                .forEach(track /*of type MediaStreamTrack*/ => track.stop()); //stop each one
        },

        /** Reset all the recording properties including the media recorder and stream being captured*/
        resetRecordingProperties: function () {
            audioRecorder.mediaRecorder = null;
            audioRecorder.streamBeingCaptured = null;

            /*No need to remove event listeners attached to mediaRecorder as
            If a DOM element which is removed is reference-free (no references pointing to it), the element itself is picked
            up by the garbage collector as well as any event handlers/listeners associated with it.
            getEventListeners(audioRecorder.mediaRecorder) will return an empty array of events.*/
        }
    }
    return audioRecorder;
}