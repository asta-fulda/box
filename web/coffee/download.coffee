###                                                                                                                                                                                                                 
 * Copyright 2011 Dustin Frisch<fooker@lab.sh>
 * 
 * This file is part of box.
 * 
 * box is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 * 
 * box is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 * 
 * You should have received a copy of the GNU General Public License
 * along with box. If not, see <http://www.gnu.org/licenses/>.
###


class Model
    constructor: () ->
        parameters = get_url_parameters()
        
        @id = parameters['i']
        @file = parameters['f']
        
        @size = ko.observable ''
        @type = ko.observable ''
        
        @loading = ko.observable true
        
        @download_url = "#{base_url}/storage/#{@id}?f=#{@file}"
        
        $.ajax
            'url': "#{base_url}/storage/#{@id}"
            'type': 'HEAD'
            'processData': false
            'success': (data, textStatus, jqXHR) =>
                @size parseInt jqXHR.getResponseHeader 'Content-Length'
                @type jqXHR.getResponseHeader 'Content-Type'
                
                console.log jqXHR.getAllResponseHeaders()
            
            'complete': () =>
                @loading(false)


ko.applyBindings new Model()
