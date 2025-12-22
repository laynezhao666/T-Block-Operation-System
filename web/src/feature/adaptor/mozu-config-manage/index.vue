<template>
  <div
    class="menu-config-page"
  >
      <el-title>模组配置管理</el-title>
      <el-block
        no-padding
        header-border
        >
       <el-table-toolbar
        hide-search
      >
        <template slot="extra">
          <el-button type="primary" @click="addMozuConfig">新增模组</el-button>
        </template>
      </el-table-toolbar> 
      <el-table
        :data="tableData"
        style="width: 100%"
        >
        <el-table-column v-for="item in tableColumns.filter(i=> i.label !== '操作')" :prop="item.props" :label="item.label" :width="item.width" :key="item.props">
          </el-table-column>
          <el-table-column prop="actions" label="操作" >
            <template slot-scope="scope">
             <el-space>
                <!-- <el-button type="text" >编辑</el-button> -->
                <el-button type="text" @click="deleteMozu(scope.row)">删除</el-button>   
             </el-space>
            </template>
          </el-table-column>
      </el-table>
    </el-block>
    <add-mozu-dialog
      :visible.sync="dialogVisible"
      @confirm="confirmMozuConfig"
    />
</div>
</template>

<script>
import AdminLimitContent from 'feature/component/tedge-components/admin-limit-content.vue';
import addMozuDialog from './components/addMozuDialog.vue';
import { tbosImportCgi } from '@@/config/cgi.js'

export default {
    components: {
        AdminLimitContent,
        addMozuDialog
    },
    data() {
    const tableColumns = [
        {
            props:'mozu_id',
            label: '模组ID'            
        },
        {
            props:'mozu_name',
            label: '模组名称',    
            width: 160       
        },
        {
            props:'mozu_code',
            label: '模组编码',    
            width: 160       
        },
        {
            props:'belong_building',
            label: '所属楼栋'            
        },
        {
            props:'belong_campus',
            label: '所属园区'            
        },
        {
            props:'belong_campus_code',
            label: '所属园区编码'            
        },
        {
            props:'publish_version',
            label: '配置版本'            
        },
        {
            props:'alarm_version',
            label: '告警版本'            
        },
        {
            props:'create_at',
            label: '创建时间'            
        },
        {
            props:'update_at',
            label: '更新时间'            
        },
        {
            props:'actions',
            label: '操作'            
        }
    ]
    return {
        tableColumns: tableColumns,
        searchValue: '',
        tableData: [],
        dialogVisible: false,
    }
    },
    mounted() {
        this.getMozuConfigData();
    },
    methods: {
        search() {
            console.log(this.searchValue);
        },
        foo() {
            console.log('bar');
        },
        async getMozuConfigData() {
            this.tableData = (await this.$axios.post(tbosImportCgi.listMozu, {}))?.list
            console.log(this.tableData,'this.tableData')
        },
        addMozuConfig(){
            this.dialogVisible = true;
        },
        deleteMozu(row){
            this.$confirm('确认删除模组？', '系统提示', { type: 'warning' }).then(() => {
                this.$axios.post(tbosImportCgi.deleteMozu, {
                    mozu_id: [row.mozu_id]
                 }).then(() => {
                this.$message.success('删除成功');
                this.getMozuConfigData();
            });
            });
        },
        confirmMozuConfig(){
            this.getMozuConfigData();
            this.dialogVisible = false;        
        },
        onSuccess({ code, data, message }) {
            console.log({ code, data, message })
        },
    }
}
  </script>