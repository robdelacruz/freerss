function targetHasClass(e, ...cc) {
    for (let i=0; i < cc.length; i++) {
        let c = cc[i];
        if (e.target.classList.contains(c)) {
            return true;
        }
    }
    return false;
}

function ondragstart(e) {
    if (!targetHasClass(e, "widget")) {
        return;
    }

    e.dataTransfer.setData("text/html", e.target.outerHTML);
}

function ondragover(e) {
    if (!targetHasClass(e, "widget", "dropzone")) {
        return;
    }

    e.preventDefault();
    e.dataTransfer.dropEffect = "move";
}

function ondrop(e) {
    if (!targetHasClass(e, "widget", "dropzone")) {
        return;
    }

    e.preventDefault();
    let outerHTML = e.dataTransfer.getData("text/html");

    if (targetHasClass(e, "widget")) {
        e.target.insertAdjacentHTML("beforebegin", outerHTML);
        return;
    }

    // "dropzone" column
    let childWidgets = e.target.querySelectorAll(".widget");
    if (childWidgets.length == 0) {
        // empty column 
        e.target.insertAdjacentHTML("afterbegin", outerHTML);
        return;
    }
    let bottomWidget = childWidgets[childWidgets.length-1];
    bottomWidget.insertAdjacentHTML("afterend", outerHTML);
}

function ondragend(e) {
    if (!targetHasClass(e, "widget", "dropzone")) {
        return;
    }

    // if drop was completed
    if (e.dataTransfer.dropEffect != "none") {
        e.target.remove();
    }
}

document.addEventListener("dragstart", ondragstart);
document.addEventListener("dragover", ondragover);
document.addEventListener("drop", ondrop);
document.addEventListener("dragend", ondragend);

