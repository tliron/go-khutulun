
const INTERVAL = 3000;

var intervals = {};
var namespace = '';

function escapeContent(html) {
  return html.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}

function escapeAttribute(html) {
  return escapeContent(html).replace(/"/g, "&quot;");
}

function syncTable(tab, url, columns) {
  let namespaced = columns.includes('namespace');
  let tabControl = $('#'+tab+'-tab');
  let headerNamespace = $('#header-namespace');
  let selectNamespace = $('#select-namespace');

  if (namespaced)
    tabControl.on('hide.bs.tab', function(event) {
      headerNamespace.addClass('invisible');
    });

  tabControl.on('show.bs.tab', function(event) {
    if (namespaced)
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
          let tr = `<tr>`
          for (let ii = 0, ll = columns.length; ii < ll; ii++) {
            let column = columns[ii];
            let value = identifier[column] || '';
            if ((column == 'namespace') && (value == '_'))
              value = '(default)'
            tr += `<td>` + escapeContent(value) + `</td>`;
          }
          tr = $(tr + `</tr>`)
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
        url: url + (namespaced ? '&namespace=' + encodeURIComponent(namespace) : ''),
        dataType: 'json',
        success: renderTable
      });
    }

    function tick() {
      if (namespaced)
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

  syncTable('services', 'api/package/list?type=service', ['namespace', 'name', 'description']);
  syncTable('templates', 'api/package/list?type=template', ['namespace', 'name', 'description']);
  syncTable('profiles', 'api/package/list?type=profile', ['namespace', 'name', 'description']);
  syncTable('delegates', 'api/package/list?type=delegate', ['namespace', 'name', 'description']);
  syncTable('activities', 'api/resource/list?type=activity', ['namespace', 'name', 'description', 'host']);
  syncTable('connections', 'api/resource/list?type=connection', ['namespace', 'name', 'description', 'host']);
  syncTable('hosts', 'api/host/list', ['name', 'grpcAddress']);

});