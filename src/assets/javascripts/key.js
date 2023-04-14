function scrollto(target, scroll) {
  var padding = 10;
  var targetRect = target.getBoundingClientRect();
  var scrollRect = scroll.getBoundingClientRect();

  // target
  var relativeOffset = targetRect.y - scrollRect.y;
  var absoluteOffset = relativeOffset + scroll.scrollTop;

  if (
    padding <= relativeOffset &&
    relativeOffset + targetRect.height <= scrollRect.height - padding
  )
    return;

  var newPos = scroll.scrollTop;
  if (relativeOffset < padding) {
    newPos = absoluteOffset - padding;
  } else {
    newPos = absoluteOffset - scrollRect.height + targetRect.height + padding;
  }
  scroll.scrollTop = Math.round(newPos);
}

var helperFunctions = {
  // navigation helper, navigate relative to selected item
  navigateToItem: function (relativePosition) {
    if (vm.itemSelected == null) {
      // if no item is selected, select first
      if (vm.items.length !== 0) vm.itemSelected = vm.items[0].id;
      return;
    }

    var itemPosition = vm.items.findIndex(function (x) {
      return x.id === vm.itemSelected;
    });
    if (itemPosition === -1) {
      if (vm.items.length !== 0) vm.itemSelected = vm.items[0].id;
      return;
    }

    var newPosition = itemPosition + relativePosition;
    if (newPosition < 0 || newPosition >= vm.items.length) return;

    vm.itemSelected = vm.items[newPosition].id;

    vm.$nextTick(function () {
      var scroll = document.querySelector("#item-list-scroll");

      var handle = scroll.querySelector("input[type=radio]:checked");
      var target = handle && handle.parentElement;

      if (target && scroll) scrollto(target, scroll);
    });
  },
  // navigation helper, navigate relative to selected feed
  navigateToFeed: function (relativePosition) {
    var navigationList = Array.from(
      document.querySelectorAll("#col-feed-list input[name=feed]")
    )
      .filter(function (r) {
        return r.offsetParent !== null && r.value !== "folder:null";
      })
      .map(function (r) {
        return r.value;
      });

    var currentFeedPosition = navigationList.indexOf(vm.feedSelected);

    if (currentFeedPosition == -1) {
      vm.feedSelected = "";
      return;
    }

    var newPosition = currentFeedPosition + relativePosition;
    if (newPosition < 0 || newPosition >= navigationList.length) return;

    vm.feedSelected = navigationList[newPosition];

    vm.$nextTick(function () {
      var scroll = document.querySelector("#feed-list-scroll");

      var handle = scroll.querySelector("input[type=radio]:checked");
      var target = handle && handle.parentElement;

      if (target && scroll) scrollto(target, scroll);
    });
  },
  scrollContent: function (direction) {
    var padding = 40;
    var scroll = document.querySelector(".content");
    if (!scroll) return;

    var height = scroll.getBoundingClientRect().height;
    var newpos = scroll.scrollTop + (height - padding) * direction;

    if (typeof scroll.scrollTo == "function") {
      scroll.scrollTo({ top: newpos, left: 0, behavior: "smooth" });
    } else {
      scroll.scrollTop = newpos;
    }
  },
};
var shortcutFunctions = {
  openItemLink: function () {
    if (vm.itemSelectedDetails && vm.itemSelectedDetails.link) {
      window.open(vm.itemSelectedDetails.link, "_blank");
    }
  },
  toggleReadability: function () {
    vm.toggleReadability();
  },
  toggleItemRead: function () {
    if (vm.itemSelected != null) {
      vm.toggleItemRead(vm.itemSelectedDetails);
    }
  },
  markAllRead: function () {
    // same condition as 'Mark all read button'
    if (vm.filterSelected == "unread") {
      vm.markItemsRead();
    }
  },
  toggleItemStarred: function () {
    if (vm.itemSelected != null) {
      vm.toggleItemStarred(vm.itemSelectedDetails);
    }
  },
  focusSearch: function () {
    document.getElementById("searchbar").focus();
  },
  nextItem() {
    helperFunctions.navigateToItem(+1);
  },
  previousItem() {
    helperFunctions.navigateToItem(-1);
  },
  nextFeed() {
    helperFunctions.navigateToFeed(+1);
  },
  previousFeed() {
    helperFunctions.navigateToFeed(-1);
  },
  scrollForward: function () {
    helperFunctions.scrollContent(+1);
  },
  scrollBackward: function () {
    helperFunctions.scrollContent(-1);
  },
  showAll() {
    vm.filterSelected = "";
  },
  showUnread() {
    vm.filterSelected = "unread";
  },
  showStarred() {
    vm.filterSelected = "starred";
  },
  toggleShortcuts() {
    vm.toggleShortcuts();
  },
  toggleFolder() {
    const current = vm.current
    if (current.type === 'folder') {
      vm.toggleFolderExpanded(current.folder)
    }
  },
  closeDialog() {
    vm.closeDialog();
  },
  toggleFullscreen() {
    const element = document.documentElement;
    const isFullscreen = element.fullscreen ||
          element.mozFullScreen ||
          element.webkitIsFullScreen ||
          element.webkitFullScreen ||
          element.msFullScreen;

    if (isFullscreen) {
      const exitMethod = document.cancelFullScreen
            || document.webkitCancelFullScreen
            || document.mozCancelFullScreen
            || document.exitFullScreen;
      if (exitMethod) {
        exitMethod.call(document)
      }
    } else {
      const requestMethod = element.requestFullScreen
            || element.webkitRequestFullScreen
            || element.mozRequestFullScreen
            || element.msRequestFullScreen;
      if (requestMethod) {
        requestMethod.call(element)
      }
    }
  }
};

// If you edit, make sure you update the help modal
var keybindings = {
  o: shortcutFunctions.openItemLink,
  i: shortcutFunctions.toggleReadability,
  r: shortcutFunctions.toggleItemRead,
  R: shortcutFunctions.markAllRead,
  y: shortcutFunctions.markAllRead,
  s: shortcutFunctions.toggleItemStarred,
  "/": shortcutFunctions.focusSearch,
  j: shortcutFunctions.nextItem,
  k: shortcutFunctions.previousItem,
  l: shortcutFunctions.nextFeed,
  h: shortcutFunctions.previousFeed,
  f: shortcutFunctions.toggleFullscreen,
  // b: shortcutFunctions.scrollBackward,
  n: shortcutFunctions.scrollForward,
  p: shortcutFunctions.scrollBackward,
  1: shortcutFunctions.showUnread,
  2: shortcutFunctions.showStarred,
  3: shortcutFunctions.showAll,
  "?": shortcutFunctions.toggleShortcuts,
};

var codebindings = {
  KeyO: shortcutFunctions.openItemLink,
  KeyI: shortcutFunctions.toggleReadability,
  //"r": shortcutFunctions.toggleItemRead,
  //"KeyR": shortcutFunctions.markAllRead,
  KeyY: shortcutFunctions.markAllRead,
  KeyS: shortcutFunctions.toggleItemStarred,
  Slash: shortcutFunctions.focusSearch,
  KeyJ: shortcutFunctions.nextItem,
  KeyK: shortcutFunctions.previousItem,
  KeyL: shortcutFunctions.nextFeed,
  KeyH: shortcutFunctions.previousFeed,
  KeyF: shortcutFunctions.toggleFullscreen,
  // KeyB: shortcutFunctions.scrollBackward,
  KeyN: shortcutFunctions.scrollForward,
  KeyP: shortcutFunctions.scrollBackward,
  Digit1: shortcutFunctions.showUnread,
  Digit2: shortcutFunctions.showStarred,
  Digit3: shortcutFunctions.showAll,
  Enter: shortcutFunctions.toggleFolder,
  Escape: shortcutFunctions.closeDialog,
};

function isTextBox(element) {
  var tagName = element.tagName.toLowerCase();
  // Input elements that aren't text
  var inputBlocklist = [
    "button",
    "checkbox",
    "color",
    "file",
    "hidden",
    "image",
    "radio",
    "range",
    "reset",
    "search",
    "submit",
  ];

  return (
    tagName === "textarea" ||
    (tagName === "input" &&
      inputBlocklist.indexOf(element.getAttribute("type").toLowerCase()) == -1)
  );
}

document.addEventListener("keydown", function (event) {
  // Ignore while focused on text or
  // when using modifier keys (to not clash with browser behaviour)
  if (
    isTextBox(event.target) ||
    event.metaKey ||
    event.ctrlKey ||
    event.altKey
  ) {
    return;
  }
  var keybindFunction = keybindings[event.key] || codebindings[event.code];
  if (keybindFunction) {
    event.preventDefault();
    keybindFunction();
  }
});
