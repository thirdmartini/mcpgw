let App = {};

App.GetTemplate = function (path, onSuccess) {
    fetch(path, {
        method: "GET",
    }).then(response => {
        return response.text();
    }).then(data => {
        //console.log("Request OK:", data);
        onSuccess(data);
    }).catch(error => {
        console.log("Request ERROR:", error);
    });
}
