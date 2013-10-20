from fluent.sender import FluentSender

sender = FluentSender('tag')
sender.emit('label', dict(a='data', b=dict(c='d')))
