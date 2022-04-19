
const INTERVAL = 3000;

var intervals = {};
var namespace = '';

function escapeContent(html) {
  return html.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}

function escapeAttribute(html) {
  return escapeContent(html).replace(/"/g, "&quot;");
}

function syncTable(tab, url) {
  let tabControl = $('#'+tab+'-tab');
  let headerNamespace = $('#header-namespace');
  let selectNamespace = $('#select-namespace');

  tabControl.on('hide.bs.tab', function(event) {
    headerNamespace.addClass('invisible');
  });

  tabControl.on('show.bs.tab', function(event) {
    headerNamespace.removeClass('invisible');
    let table = $('#'+tab+' table');
    let tbody = $('#'+tab+' table tbody');

    function renderTable(identifiers) {
      if (!identifiers) identifiers = [];
      let exists = [];

      if (namespace == '')
        table.removeClass('hide-first-column');
      else
        table.addClass('hide-first-column');

      function find(identifier, identifiers) {
        for (let i = 0, l = identifiers.length; i < l; i++) {
          let identifier_ = identifiers[i];
          if ((identifier.namespace == identifier_.namespace) && (identifier.name == identifier_.name))
            return true;
        }
        return false;
      }

      tbody.children('tr').each(function () {
        let tr = $(this);
        let identifier = tr.data('identifier');
        exists.push(identifier);
        if (find(identifier, identifiers))
          exists.push(identifier);
        else
          tr.remove();
      });

      for (let i = 0, l = identifiers.length; i < l; i++) {
        let identifier = identifiers[i];
        if (!find(identifier, exists)) {
          let namespace_ = identifier.namespace == '_' ? '(default)' : identifier.namespace;
          let tr = $(`<tr><td>` + escapeContent(namespace_) + `</td><td>` + escapeContent(identifier.name) + `</td><td></td></tr>`);
          tr.data('identifier', identifier);
          tbody.append(tr);
        }
      }
    }

    function onSelectNamespaceChanged() {
      namespace = $(this).val();
      tickTable();
    }

    function renderNamespaces(namespaces) {
      selectNamespace.off('change');
      selectNamespace.empty();
      selectNamespace.append(`<option value="">(all)</option>`);
      for (let i = 0, l = namespaces.length; i < l; i++) {
        let namespace_ = namespaces[i];
        let text = namespace_ == '_' ? '(default)' : namespace_;
        selectNamespace.append(`<option ` + (namespace == namespace_ ? `selected ` : ` `) + `value="` + escapeAttribute(namespace_) + `">` + escapeContent(text) + `</option>`);
      }
      selectNamespace.change(onSelectNamespaceChanged);
    }

    function tickNamespaces() {
      $.get({
        url: 'api/namespace/list',
        dataType: 'json',
        success: renderNamespaces
      });
    }

    function tickTable() {
      $.get({
        url: url + '&namespace=' + encodeURIComponent(namespace),
        dataType: 'json',
        success: renderTable
      });
    }

    function tick() {
      tickNamespaces();
      tickTable();
    }

    tick();
    intervals[tab] = setInterval(tick, INTERVAL);
  });

  $('#'+tab+'-tab').on('hide.bs.tab', function(event) {
    selectNamespace.off('change');
    clearInterval(intervals[tab]);
  });
}

$(document).ready(function () {

  $('#tab button').each(function () {
    let tab = new bootstrap.Tab(this);
    $(this).on('click', function () {
      tab.show();
    });
  });

  syncTable('services', 'api/bundle/list?type=clout');
  syncTable('templates', 'api/bundle/list?type=template');
  syncTable('profiles', 'api/bundle/list?type=profile');
  syncTable('plugins', 'api/bundle/list?type=plugin');
  syncTable('runnables', 'api/resource/list?type=runnable');

});